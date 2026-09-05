package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Routes(w Wires) *gin.Engine {
	r := gin.Default()

	r.GET("/_/health", handleHealth)

	api := r.Group("/api")
	{
		api.GET("/lines", handleGetLines(w))
		api.GET("/lines/:line_id/repeats", handleGetLineRepeats(w))
		// Query param avoids path-matching issues with geometry_key underscores
		api.GET("/lines/:line_id/group-readings", handleGetGroupReadings(w))
		api.GET("/readings/:id", handleGetReadingDetail(w))
		api.POST("/readings/:id/qc", handleUpdateQC(w))
	}

	fileServer := http.FileServer(http.Dir(w.WebRoot))
	r.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	return r
}

func handleHealth(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

// 1. Line Summary Stats (Level 1)
func handleGetLines(w Wires) gin.HandlerFunc {
	type lineSummary struct {
		LineID      string  `json:"line_id"`
		TotalCount  int     `json:"total_count"`
		PassCount   int     `json:"pass_count"`
		FlagCount   int     `json:"flag_count"`
		RejectCount int     `json:"reject_count"`
		Unreviewed  int     `json:"unreviewed_count"`
		AvgRes      float64 `json:"avg_apparent_resistivity"`
		AvgCharg    float64 `json:"avg_chargeability"`
	}

	return func(ctx *gin.Context) {
		query := `
			SELECT 
				line_id,
				COUNT(*) as total_count,
				SUM(CASE WHEN qc_review = 'pass' THEN 1 ELSE 0 END) as pass_count,
				SUM(CASE WHEN qc_review = 'flag' THEN 1 ELSE 0 END) as flag_count,
				SUM(CASE WHEN qc_review = 'reject' THEN 1 ELSE 0 END) as reject_count,
				SUM(CASE WHEN qc_review IS NULL OR qc_review = '' THEN 1 ELSE 0 END) as unreviewed_count,
				AVG(apparent_resistivity) as avg_res,
				AVG(chargeability) as avg_charg
			FROM readings
			GROUP BY line_id
			ORDER BY line_id ASC;
		`

		rows, err := w.DB.Read.QueryContext(ctx.Request.Context(), query)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch line summaries"})
			return
		}
		defer rows.Close()

		var summaries []lineSummary
		for rows.Next() {
			var s lineSummary
			if err := rows.Scan(&s.LineID, &s.TotalCount, &s.PassCount, &s.FlagCount, &s.RejectCount, &s.Unreviewed, &s.AvgRes, &s.AvgCharg); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan row"})
				return
			}
			summaries = append(summaries, s)
		}

		ctx.JSON(http.StatusOK, summaries)
	}
}

// 2. Repeat Geometry Aggregations for a Line (Level 2)
func handleGetLineRepeats(w Wires) gin.HandlerFunc {
	type repeatGroup struct {
		GeometryKey string   `json:"geometry_key"`
		Tx1ID       string   `json:"tx1_id"`
		Tx2ID       string   `json:"tx2_id"`
		Rx1ID       string   `json:"rx1_id"`
		Rx2ID       string   `json:"rx2_id"`
		Count       int      `json:"repeat_count"`
		AvgCharg    float64  `json:"avg_chargeability"`
		VarCharg    float64  `json:"var_chargeability"`
		AvgRes      float64  `json:"avg_apparent_resistivity"`
		VarRes      float64  `json:"var_apparent_resistivity"`
		ReadingIDs  []string `json:"reading_ids"`
	}

	return func(ctx *gin.Context) {
		lineID := ctx.Param("line_id")

		query := `
			SELECT 
				geometry_key,
				tx1_id, tx2_id, rx1_id, rx2_id,
				COUNT(*) as repeat_count,
				AVG(chargeability) as avg_charg,
				COALESCE(AVG((chargeability - sub.avg_c) * (chargeability - sub.avg_c)), 0) as var_charg,
				AVG(apparent_resistivity) as avg_res,
				COALESCE(AVG((apparent_resistivity - sub.avg_r) * (apparent_resistivity - sub.avg_r)), 0) as var_res,
				GROUP_CONCAT(id) as reading_ids
			FROM readings
			JOIN (
				SELECT geometry_key as gk, AVG(chargeability) as avg_c, AVG(apparent_resistivity) as avg_r
				FROM readings WHERE line_id = ? GROUP BY geometry_key
			) sub ON readings.geometry_key = sub.gk
			WHERE line_id = ?
			GROUP BY geometry_key
			HAVING repeat_count > 1
			ORDER BY repeat_count DESC, var_charg DESC;
		`

		rows, err := w.DB.Read.QueryContext(ctx.Request.Context(), query, lineID, lineID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query repeat groups"})
			return
		}
		defer rows.Close()

		groups := []repeatGroup{}
		for rows.Next() {
			var g repeatGroup
			var idsCSV string
			if err := rows.Scan(&g.GeometryKey, &g.Tx1ID, &g.Tx2ID, &g.Rx1ID, &g.Rx2ID, &g.Count, &g.AvgCharg, &g.VarCharg, &g.AvgRes, &g.VarRes, &idsCSV); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan repeat group"})
				return
			}
			if idsCSV != "" {
				g.ReadingIDs = strings.Split(idsCSV, ",")
			} else {
				g.ReadingIDs = []string{}
			}
			groups = append(groups, g)
		}

		ctx.JSON(http.StatusOK, groups)
	}
}

// 2b. All readings for one repeat geometry (decay curves for Level 3)
func handleGetGroupReadings(w Wires) gin.HandlerFunc {
	type readingDetail struct {
		ID                  string  `json:"id"`
		LineID              string  `json:"line_id"`
		Tx1ID               string  `json:"tx1_id"`
		Tx2ID               string  `json:"tx2_id"`
		Rx1ID               string  `json:"rx1_id"`
		Rx2ID               string  `json:"rx2_id"`
		Timestamp           string  `json:"timestamp"`
		ApparentResistivity float64 `json:"apparent_resistivity"`
		ResistivityErr      float64 `json:"apparent_resistivity_err"`
		Chargeability       float64 `json:"chargeability"`
		ChargeabilityErr    float64 `json:"chargeability_err"`
		DecayCurve          string  `json:"decay_curve"`
		QCReview            string  `json:"qc_review"`
	}

	return func(ctx *gin.Context) {
		lineID := ctx.Param("line_id")
		geometryKey := ctx.Query("geometry_key")
		if geometryKey == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "geometry_key query param is required"})
			return
		}

		query := `
			SELECT id, line_id, tx1_id, tx2_id, rx1_id, rx2_id, timestamp,
			       apparent_resistivity, apparent_resistivity_err,
			       chargeability, chargeability_err, decay_curve, COALESCE(qc_review, '')
			FROM readings
			WHERE line_id = ? AND geometry_key = ?
			ORDER BY timestamp ASC
		`

		rows, err := w.DB.Read.QueryContext(ctx.Request.Context(), query, lineID, geometryKey)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query group readings"})
			return
		}
		defer rows.Close()

		readings := []readingDetail{}
		for rows.Next() {
			var r readingDetail
			if err := rows.Scan(
				&r.ID, &r.LineID, &r.Tx1ID, &r.Tx2ID, &r.Rx1ID, &r.Rx2ID, &r.Timestamp,
				&r.ApparentResistivity, &r.ResistivityErr,
				&r.Chargeability, &r.ChargeabilityErr, &r.DecayCurve, &r.QCReview,
			); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan reading"})
				return
			}
			readings = append(readings, r)
		}

		ctx.JSON(http.StatusOK, readings)
	}
}

// 3. Single Reading Detail + Decay Curve (Level 3)
func handleGetReadingDetail(w Wires) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		query := `
			SELECT id, line_id, tx1_id, tx2_id, rx1_id, rx2_id, timestamp, 
			       apparent_resistivity, apparent_resistivity_err, 
			       chargeability, chargeability_err, decay_curve, COALESCE(qc_review, '')
			FROM readings WHERE id = ?
		`

		var r struct {
			ID                  string  `json:"id"`
			LineID              string  `json:"line_id"`
			Tx1ID               string  `json:"tx1_id"`
			Tx2ID               string  `json:"tx2_id"`
			Rx1ID               string  `json:"rx1_id"`
			Rx2ID               string  `json:"rx2_id"`
			Timestamp           string  `json:"timestamp"`
			ApparentResistivity float64 `json:"apparent_resistivity"`
			ResistivityErr      float64 `json:"apparent_resistivity_err"`
			Chargeability       float64 `json:"chargeability"`
			ChargeabilityErr    float64 `json:"chargeability_err"`
			DecayCurve          string  `json:"decay_curve"`
			QCReview            string  `json:"qc_review"`
		}

		err := w.DB.Read.QueryRowContext(ctx.Request.Context(), query, id).Scan(
			&r.ID, &r.LineID, &r.Tx1ID, &r.Tx2ID, &r.Rx1ID, &r.Rx2ID, &r.Timestamp,
			&r.ApparentResistivity, &r.ResistivityErr,
			&r.Chargeability, &r.ChargeabilityErr, &r.DecayCurve, &r.QCReview,
		)

		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "reading not found"})
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		ctx.JSON(http.StatusOK, r)
	}
}

// 4. Update QC Status ('pass', 'flag', 'reject')
func handleUpdateQC(w Wires) gin.HandlerFunc {
	type qcReq struct {
		Status string `json:"status" binding:"required"`
	}

	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		var req qcReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}

		status := strings.ToLower(req.Status)
		if status != "pass" && status != "flag" && status != "reject" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'pass', 'flag', or 'reject'"})
			return
		}

		res, err := w.DB.Write.ExecContext(ctx.Request.Context(), "UPDATE readings SET qc_review = ? WHERE id = ?", status, id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update review status"})
			return
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "reading not found"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"id": id, "status": status})
	}
}