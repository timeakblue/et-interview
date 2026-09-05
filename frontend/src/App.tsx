import React, { useEffect, useMemo, useState } from 'react'
import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Paper from '@mui/material/Paper'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Typography from '@mui/material/Typography'
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  CheckCircle,
  Database,
  RefreshCw,
  XCircle,
} from 'lucide-react'
import { colors } from './theme'

// ----------------------------------------------------------------------
// Interfaces
// ----------------------------------------------------------------------
interface LineSummary {
  line_id: string
  total_count: number
  pass_count: number
  flag_count: number
  reject_count: number
  unreviewed_count: number
  avg_apparent_resistivity: number
  avg_chargeability: number
}

interface RepeatGroup {
  geometry_key: string
  tx1_id: string
  tx2_id: string
  rx1_id: string
  rx2_id: string
  repeat_count: number
  avg_chargeability: number
  var_chargeability: number
  avg_apparent_resistivity: number
  var_apparent_resistivity: number
  reading_ids?: string[]
}

interface ReadingDetail {
  id: string
  line_id: string
  tx1_id: string
  tx2_id: string
  rx1_id: string
  rx2_id: string
  timestamp: string
  apparent_resistivity: number
  apparent_resistivity_err: number
  chargeability: number
  chargeability_err: number
  decay_curve: string
  qc_review: string
}

// ----------------------------------------------------------------------
// Helper: Parse Semicolon/Space/Comma-Separated decay_curve String
// ----------------------------------------------------------------------
// Gate centres from the survey spec (ms)
export const DECAY_GATE_MS = [24.32, 35.96, 53.18, 78.64, 116.3, 171.97, 254.31, 376.06, 556.1, 822.34]

export const parseDecayCurve = (decayString?: string) => {
  if (!decayString) return []

  return decayString
    .split(/[\s,;]+/)
    .map((val) => parseFloat(val.trim()))
    .filter((val) => !isNaN(val))
    .map((value, index) => ({
      gate: index + 1,
      timeMs: DECAY_GATE_MS[index] ?? index + 1,
      // Values are already V/V; scale to mV/V for readable plotting when small
      value: value > 1 ? value : value * 1000,
    }))
}

// ----------------------------------------------------------------------
// Level 3 Component: Single Reading Decay Curve Plot
// ----------------------------------------------------------------------
function DecayCurvePlot({ decayString, title }: { decayString?: string; title?: string }) {
  const chartData = useMemo(() => parseDecayCurve(decayString), [decayString])

  if (chartData.length === 0) {
    return (
      <Paper variant="outlined" sx={{ p: 2, borderColor: colors.border, bgcolor: 'background.paper' }}>
        <Typography sx={{ fontSize: '12px', color: colors.textSecondary, py: 2, textAlign: 'center' }}>
          No decay curve data available
        </Typography>
      </Paper>
    )
  }

  // Detect non-monotonic gates (chargeability decay should typically decrease over time)
  const nonMonotonicIndices: number[] = []
  for (let i = 1; i < chartData.length; i++) {
    if (chartData[i].value > chartData[i - 1].value) {
      nonMonotonicIndices.push(chartData[i].gate)
    }
  }

  const svgWidth = 320
  const svgHeight = 160
  const padding = 28

  const yValues = chartData.map((d) => d.value)
  const minY = Math.min(...yValues, 0)
  const maxY = Math.max(...yValues, 10)

  const points = chartData.map((d, i) => {
    const x = padding + (i / (chartData.length - 1 || 1)) * (svgWidth - 2 * padding)
    const y = svgHeight - padding - ((d.value - minY) / (maxY - minY || 1)) * (svgHeight - 2 * padding)
    return { x, y, ...d }
  })

  const pathD = points.reduce((acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`), '')

  return (
    <Paper variant="outlined" sx={{ p: 2, borderColor: colors.border, bgcolor: 'background.paper' }}>
      <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
        <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary }}>
          {title ?? `Decay Curve (${chartData.length} Gates)`}
        </Typography>
        {nonMonotonicIndices.length > 0 && (
          <Typography sx={{ fontSize: '11px', fontWeight: 600, color: colors.error }}>
            Non-monotonic ({nonMonotonicIndices.length})
          </Typography>
        )}
      </Stack>

      <Box sx={{ border: `1px solid ${colors.border}`, borderRadius: '4px', bgcolor: '#fafafa', p: 1 }}>
        <svg width="100%" height={svgHeight} viewBox={`0 0 ${svgWidth} ${svgHeight}`}>
          <line x1={padding} y1={svgHeight - padding} x2={svgWidth - padding} y2={svgHeight - padding} stroke="#e2e8f0" strokeWidth="1" />
          <line x1={padding} y1={padding} x2={padding} y2={svgHeight - padding} stroke="#e2e8f0" strokeWidth="1" />

          <path d={pathD} fill="none" stroke="#2563eb" strokeWidth="2" />

          {points.map((p) => {
            const isBad = nonMonotonicIndices.includes(p.gate)
            return (
              <circle
                key={p.gate}
                cx={p.x}
                cy={p.y}
                r={isBad ? 5 : 3.5}
                fill={isBad ? colors.error : '#2563eb'}
                stroke="#ffffff"
                strokeWidth="1"
              >
                <title>{`Gate ${p.gate} @ ${p.timeMs} ms: ${p.value.toFixed(2)} mV/V`}</title>
              </circle>
            )
          })}

          {/* First / last time labels */}
          <text x={padding} y={svgHeight - 8} fontSize="9" fill="#64748b">
            {chartData[0].timeMs} ms
          </text>
          <text x={svgWidth - padding} y={svgHeight - 8} fontSize="9" fill="#64748b" textAnchor="end">
            {chartData[chartData.length - 1].timeMs} ms
          </text>
        </svg>
      </Box>

      {nonMonotonicIndices.length > 0 ? (
        <Typography sx={{ fontSize: '11px', color: colors.error, mt: 1, fontWeight: 500 }}>
          Anomaly at gate(s): {nonMonotonicIndices.join(', ')}
        </Typography>
      ) : (
        <Typography sx={{ fontSize: '11px', color: colors.success, mt: 1, fontWeight: 500 }}>
          Regular decay toward zero
        </Typography>
      )}
    </Paper>
  )
}

// ----------------------------------------------------------------------
// Lightweight Inline SVG Component: Level 2 Repeat Variance Overlay
// ----------------------------------------------------------------------
const LINE_COLORS = ['#2563eb', '#dc2626', '#16a34a', '#ca8a04', '#9333ea', '#0891b2', '#ea580c', '#4f46e5']

function RepeatVariancePlot({
  selectedGroup,
  groupReadings,
  selectedReadingId,
  onSelectReading,
}: {
  selectedGroup: RepeatGroup
  groupReadings: ReadingDetail[]
  selectedReadingId: string | null
  onSelectReading: (id: string) => void
}) {
  if (groupReadings.length === 0) return null

  const svgWidth = 500
  const svgHeight = 180
  const padding = 28

  const allParsed = groupReadings.map((r) => parseDecayCurve(r.decay_curve))
  const allValues = allParsed.flatMap((curve) => curve.map((c) => c.value))

  if (allValues.length === 0) return null

  const minY = Math.min(...allValues, 0)
  const maxY = Math.max(...allValues, 10)

  return (
    <Box sx={{ bgcolor: '#ffffff', p: 2, mb: 2, borderRadius: '6px', border: `1px solid ${colors.border}` }}>
      <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, mb: 1 }}>
        Repeat Decay Overlay — {selectedGroup.geometry_key} ({groupReadings.length} repeats)
      </Typography>

      <svg width="100%" height={svgHeight} viewBox={`0 0 ${svgWidth} ${svgHeight}`}>
        <line x1={padding} y1={svgHeight - padding} x2={svgWidth - padding} y2={svgHeight - padding} stroke="#e2e8f0" strokeWidth="1" />
        <line x1={padding} y1={padding} x2={padding} y2={svgHeight - padding} stroke="#e2e8f0" strokeWidth="1" />

        {allParsed.map((curve, idx) => {
          if (curve.length === 0) return null
          const reading = groupReadings[idx]
          const color = LINE_COLORS[idx % LINE_COLORS.length]
          const isSelected = selectedReadingId === reading.id

          const points = curve.map((d, i) => {
            const x = padding + (i / (curve.length - 1 || 1)) * (svgWidth - 2 * padding)
            const y = svgHeight - padding - ((d.value - minY) / (maxY - minY || 1)) * (svgHeight - 2 * padding)
            return { x, y, ...d }
          })

          const pathD = points.reduce((acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`), '')

          return (
            <g
              key={reading.id}
              style={{ cursor: 'pointer' }}
              onClick={() => onSelectReading(reading.id)}
              opacity={selectedReadingId && !isSelected ? 0.35 : 0.95}
            >
              <path d={pathD} fill="none" stroke={color} strokeWidth={isSelected ? 3 : 2} />
              {points.map((p) => (
                <circle key={p.gate} cx={p.x} cy={p.y} r={isSelected ? 4 : 2.5} fill={color}>
                  <title>{`${reading.id} — Gate ${p.gate} @ ${p.timeMs} ms: ${p.value.toFixed(2)} mV/V`}</title>
                </circle>
              ))}
            </g>
          )
        })}
      </svg>

      <Stack direction="row" flexWrap="wrap" gap={1} sx={{ mt: 1 }}>
        {groupReadings.map((r, idx) => (
          <Button
            key={r.id}
            size="small"
            variant={selectedReadingId === r.id ? 'contained' : 'outlined'}
            onClick={() => onSelectReading(r.id)}
            sx={{
              textTransform: 'none',
              fontSize: '11px',
              py: 0.25,
              px: 1,
              minWidth: 0,
              borderColor: LINE_COLORS[idx % LINE_COLORS.length],
              color: selectedReadingId === r.id ? '#fff' : LINE_COLORS[idx % LINE_COLORS.length],
              bgcolor: selectedReadingId === r.id ? LINE_COLORS[idx % LINE_COLORS.length] : 'transparent',
              '&:hover': {
                borderColor: LINE_COLORS[idx % LINE_COLORS.length],
                bgcolor: selectedReadingId === r.id ? LINE_COLORS[idx % LINE_COLORS.length] : `${LINE_COLORS[idx % LINE_COLORS.length]}14`,
              },
            }}
          >
            {r.id.replace(/^.*-r/, 'r')}
          </Button>
        ))}
      </Stack>
    </Box>
  )
}

// ----------------------------------------------------------------------
// Main Application Component
// ----------------------------------------------------------------------
export default function App() {
  const [lines, setLines] = useState<LineSummary[]>([])
  const [selectedLine, setSelectedLine] = useState<string>('')
  const [repeats, setRepeats] = useState<RepeatGroup[]>([])
  const [selectedGroup, setSelectedGroup] = useState<RepeatGroup | null>(null)
  const [selectedReadingId, setSelectedReadingId] = useState<string | null>(null)
  const [groupReadings, setGroupReadings] = useState<ReadingDetail[]>([])
  const [groupLoading, setGroupLoading] = useState(false)
  const [groupError, setGroupError] = useState<string | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  const readingDetail = useMemo(() => {
    if (!selectedReadingId) return groupReadings[0] ?? null
    return groupReadings.find((r) => r.id === selectedReadingId) ?? groupReadings[0] ?? null
  }, [groupReadings, selectedReadingId])

  const fetchLines = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await fetch('/api/lines')
      if (!res.ok) throw new Error('Failed to load line summaries')
      const data = await res.json()
      setLines(data || [])
      if (data && data.length > 0 && !selectedLine) {
        setSelectedLine(data[0].line_id)
      }
    } catch (err: any) {
      setError(err.message || 'Error communicating with server')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLines()
  }, [])

  // Fetch repeat geometries when line changes
  useEffect(() => {
    if (!selectedLine) return
    let cancelled = false

    const fetchRepeats = async () => {
      try {
        const res = await fetch(`/api/lines/${selectedLine}/repeats`)
        if (!res.ok) throw new Error('Failed to load repeat geometries')
        const data: RepeatGroup[] = await res.json()
        if (cancelled) return
        setRepeats(data || [])
        if (data && data.length > 0) {
          setSelectedGroup(data[0])
          setSelectedReadingId(data[0].reading_ids?.[0] ?? null)
        } else {
          setSelectedGroup(null)
          setSelectedReadingId(null)
          setGroupReadings([])
        }
      } catch (err) {
        console.error(err)
      }
    }
    fetchRepeats()
    return () => {
      cancelled = true
    }
  }, [selectedLine])

  // Fetch all decay curves for the selected geometry group
  useEffect(() => {
    if (!selectedLine || !selectedGroup) {
      setGroupReadings([])
      return
    }

    let cancelled = false
    setGroupLoading(true)
    setGroupError(null)

    const fetchGroup = async () => {
      try {
        // Prefer one-shot group endpoint (query param — no path encoding issues)
        const key = encodeURIComponent(selectedGroup.geometry_key)
        const groupUrl = `/api/lines/${selectedLine}/group-readings?geometry_key=${key}`
        let data: ReadingDetail[] | null = null

        const groupRes = await fetch(groupUrl)
        if (groupRes.ok) {
          data = await groupRes.json()
        } else if (selectedGroup.reading_ids && selectedGroup.reading_ids.length > 0) {
          // Fallback: fetch each reading via the existing detail endpoint
          const results = await Promise.all(
            selectedGroup.reading_ids.map(async (id) => {
              const res = await fetch(`/api/readings/${encodeURIComponent(id)}`)
              if (!res.ok) throw new Error(`reading ${id}: HTTP ${res.status}`)
              return res.json() as Promise<ReadingDetail>
            })
          )
          data = results
        } else {
          throw new Error(
            `Group readings 404 — restart the Go backend (go run .) so /group-readings and reading_ids are available`
          )
        }

        if (cancelled || !data) return
        setGroupReadings(data)
        setSelectedReadingId((prev) => {
          if (prev && data!.some((r) => r.id === prev)) return prev
          return data![0]?.id ?? null
        })
      } catch (err: any) {
        if (!cancelled) {
          console.error(err)
          setGroupError(err.message || 'Failed to load decay curves')
          setGroupReadings([])
        }
      } finally {
        if (!cancelled) setGroupLoading(false)
      }
    }

    fetchGroup()
    return () => {
      cancelled = true
    }
  }, [selectedLine, selectedGroup])

  const updateQCStatus = async (id: string, status: 'pass' | 'flag' | 'reject') => {
    try {
      const res = await fetch(`/api/readings/${id}/qc`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status }),
      })
      if (res.ok) {
        setGroupReadings((prev) => prev.map((r) => (r.id === id ? { ...r, qc_review: status } : r)))
        fetchLines()
      }
    } catch (err) {
      console.error(err)
    }
  }

  const totalReadings = lines.reduce((acc, l) => acc + l.total_count, 0)
  const totalPassed = lines.reduce((acc, l) => acc + l.pass_count, 0)
  const totalFlagged = lines.reduce((acc, l) => acc + l.flag_count, 0)
  const totalRejected = lines.reduce((acc, l) => acc + l.reject_count, 0)
  const totalUnreviewed = lines.reduce((acc, l) => acc + l.unreviewed_count, 0)

  return (
    <Box sx={{ height: '100vh', width: '100vw', overflow: 'hidden', display: 'flex', bgcolor: 'background.paper' }}>
      {/* 52px Side Rail */}
      <Box
        sx={{
          width: 52,
          bgcolor: colors.primary,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          pt: 2,
          zIndex: 40,
          flexShrink: 0,
        }}
      >
        <Box sx={{ color: '#ffffff', mb: 3, opacity: 0.9 }}>
          <Database size={22} strokeWidth={1.5} />
        </Box>
        <Box
          sx={{
            width: '100%',
            display: 'flex',
            justifyContent: 'center',
            py: 1,
            borderLeft: '2px solid #ffffff',
            color: '#ffffff',
            cursor: 'pointer',
          }}
        >
          <Activity size={20} strokeWidth={1.5} />
        </Box>
      </Box>

      {/* Main Container */}
      <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {/* Header */}
        <Box
          component="header"
          sx={{
            px: 3,
            py: 2,
            borderBottom: `1px solid ${colors.border}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Box>
            <Typography variant="h1" sx={{ fontSize: '18px', fontWeight: 600, color: colors.textPrimary }}>
              Resistivity & IP Survey QC Dashboard
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, fontSize: '13px' }}>
              Electrode Geometry Variance & Quality Control Inspector
            </Typography>
          </Box>
          <Button
            variant="outlined"
            onClick={fetchLines}
            disabled={loading}
            startIcon={<RefreshCw size={14} strokeWidth={1.5} />}
            sx={{
              borderColor: colors.border,
              color: colors.textPrimary,
              '&:hover': { borderColor: colors.border, bgcolor: colors.surface },
              fontSize: '13px',
              textTransform: 'none',
            }}
          >
            Refresh
          </Button>
        </Box>

        {/* Level 1 Global Ribbon */}
        <Box
          sx={{
            display: 'grid',
            gridTemplateColumns: 'repeat(5, 1fr)',
            gap: 2,
            p: 2,
            px: 3,
            bgcolor: colors.surface,
            borderBottom: `1px solid ${colors.border}`,
          }}
        >
          <Paper variant="outlined" sx={{ p: 1.5, borderColor: colors.border }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, letterSpacing: '0.05em' }}>
              Total Readings
            </Typography>
            <Typography sx={{ fontSize: '20px', fontWeight: 600, color: colors.textPrimary, mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {totalReadings.toLocaleString()}
            </Typography>
          </Paper>
          <Paper variant="outlined" sx={{ p: 1.5, borderColor: colors.border }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, letterSpacing: '0.05em' }}>
              Unreviewed
            </Typography>
            <Typography sx={{ fontSize: '20px', fontWeight: 600, color: '#2563eb', mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {totalUnreviewed.toLocaleString()}
            </Typography>
          </Paper>
          <Paper variant="outlined" sx={{ p: 1.5, borderColor: colors.border }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.success, letterSpacing: '0.05em' }}>
              Passed
            </Typography>
            <Typography sx={{ fontSize: '20px', fontWeight: 600, color: colors.success, mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {totalPassed.toLocaleString()}
            </Typography>
          </Paper>
          <Paper variant="outlined" sx={{ p: 1.5, borderColor: colors.border }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.warning, letterSpacing: '0.05em' }}>
              Flagged
            </Typography>
            <Typography sx={{ fontSize: '20px', fontWeight: 600, color: colors.warning, mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {totalFlagged.toLocaleString()}
            </Typography>
          </Paper>
          <Paper variant="outlined" sx={{ p: 1.5, borderColor: colors.border }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.error, letterSpacing: '0.05em' }}>
              Rejected
            </Typography>
            <Typography sx={{ fontSize: '20px', fontWeight: 600, color: colors.error, mt: 0.5, fontVariantNumeric: 'tabular-nums' }}>
              {totalRejected.toLocaleString()}
            </Typography>
          </Paper>
        </Box>

        {/* Line Tabs */}
        <Box sx={{ display: 'flex', borderBottom: `1px solid ${colors.border}`, px: 3, bgcolor: 'background.paper' }}>
          {lines.map((line) => (
            <Button
              key={line.line_id}
              onClick={() => setSelectedLine(line.line_id)}
              sx={{
                py: 1.5,
                px: 2.5,
                fontSize: '13px',
                fontWeight: selectedLine === line.line_id ? 600 : 400,
                color: selectedLine === line.line_id ? colors.primary : colors.textSecondary,
                borderBottom: selectedLine === line.line_id ? `2px solid ${colors.primary}` : '2px solid transparent',
                borderRadius: 0,
                textTransform: 'none',
              }}
            >
              {line.line_id.toUpperCase()} ({line.total_count})
            </Button>
          ))}
        </Box>

        {/* Level 2 Table & Level 3 Inspector Panel */}
        <Box sx={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {/* Level 2 Area */}
          <Box sx={{ flex: 1, p: 2.5, overflowY: 'auto' }}>
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, letterSpacing: '0.05em', mb: 0.5 }}>
              Level 2: Repeat Electrode Geometries ({selectedLine.toUpperCase()})
            </Typography>
            <Typography sx={{ fontSize: '12px', color: colors.textSecondary, mb: 2 }}>
              Electrode pairs with repeated measurements sorted by variance
            </Typography>

            {/* Level 2 Variance Chart */}
            {selectedGroup && (
              <RepeatVariancePlot
                selectedGroup={selectedGroup}
                groupReadings={groupReadings}
                selectedReadingId={selectedReadingId}
                onSelectReading={setSelectedReadingId}
              />
            )}

            {loading ? (
              <Skeleton variant="rounded" animation="wave" height={240} />
            ) : error ? (
              <Alert severity="error" icon={<AlertCircle size={16} />}>
                {error}
              </Alert>
            ) : repeats.length === 0 ? (
              <Typography sx={{ fontSize: '13px', color: colors.textSecondary, py: 4, textAlign: 'center' }}>
                No repeated geometry setups found for this survey line.
              </Typography>
            ) : (
              <Paper variant="outlined" sx={{ borderColor: colors.border }}>
                <Table size="small">
                  <TableHead sx={{ bgcolor: colors.surface }}>
                    <TableRow>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Geometry Key</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Tx Pair</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Rx Pair</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Repeats</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Avg Res (Ωm)</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Avg Charg (mV/V)</TableCell>
                      <TableCell sx={{ fontSize: '11px', fontWeight: 600, color: colors.textSecondary, textTransform: 'uppercase' }}>Charg Var</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {repeats.map((g) => {
                      const isHighVar = g.var_chargeability > 0.0001
                      const isSelected = selectedGroup?.geometry_key === g.geometry_key

                      return (
                        <TableRow
                          key={g.geometry_key}
                          hover
                          selected={isSelected}
                          onClick={() => {
                            setSelectedGroup(g)
                            setSelectedReadingId(g.reading_ids?.[0] ?? null)
                          }}
                          sx={{ cursor: 'pointer' }}
                        >
                          <TableCell sx={{ fontFamily: 'monospace', fontSize: '12px' }}>{g.geometry_key}</TableCell>
                          <TableCell sx={{ fontSize: '13px' }}>{g.tx1_id} / {g.tx2_id}</TableCell>
                          <TableCell sx={{ fontSize: '13px' }}>{g.rx1_id} / {g.rx2_id}</TableCell>
                          <TableCell sx={{ fontSize: '13px', fontWeight: 500 }}>{g.repeat_count}</TableCell>
                          <TableCell sx={{ fontSize: '13px' }}>{g.avg_apparent_resistivity.toFixed(2)}</TableCell>
                          <TableCell sx={{ fontSize: '13px' }}>{g.avg_chargeability.toFixed(4)}</TableCell>
                          <TableCell sx={{ fontSize: '13px', color: isHighVar ? colors.error : colors.success, fontWeight: isHighVar ? 600 : 400 }}>
                            {g.var_chargeability.toFixed(6)} {isHighVar && '⚠️'}
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </Paper>
            )}
          </Box>

          {/* Level 3: Reading Inspector Side Panel */}
          <Box
            sx={{
              width: 380,
              borderLeft: `1px solid ${colors.border}`,
              p: 2.5,
              bgcolor: colors.surface,
              display: 'flex',
              flexDirection: 'column',
              gap: 2,
              overflowY: 'auto',
            }}
          >
            <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, letterSpacing: '0.05em' }}>
              Level 3: Reading Inspector
            </Typography>

            {groupLoading ? (
              <Box sx={{ py: 4, textAlign: 'center' }}>
                <Skeleton variant="rounded" height={120} sx={{ mb: 2 }} />
                <Typography variant="body2" color="text.secondary">
                  Loading decay curves…
                </Typography>
              </Box>
            ) : groupError ? (
              <Alert severity="error" icon={<AlertCircle size={16} />}>
                {groupError}
              </Alert>
            ) : readingDetail ? (
              <>
                <Paper variant="outlined" sx={{ p: 2, borderColor: colors.border, bgcolor: 'background.paper' }}>
                  <Typography sx={{ fontSize: '14px', fontWeight: 600, color: colors.textPrimary }}>
                    Reading #{readingDetail.id}
                  </Typography>
                  <Typography sx={{ fontSize: '12px', color: colors.textSecondary, mt: 0.5 }}>
                    Timestamp: {readingDetail.timestamp || 'N/A'}
                  </Typography>
                  {selectedGroup && (
                    <Typography sx={{ fontSize: '11px', color: colors.textSecondary, mt: 0.5, fontFamily: 'monospace' }}>
                      {selectedGroup.geometry_key}
                    </Typography>
                  )}

                  <Box sx={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 1, mt: 1.5 }}>
                    <Box>
                      <Typography sx={{ fontSize: '11px', color: colors.textSecondary }}>App. Resistivity</Typography>
                      <Typography sx={{ fontSize: '13px', fontWeight: 600 }}>
                        {readingDetail.apparent_resistivity?.toFixed(2) ?? 'N/A'} Ωm
                      </Typography>
                    </Box>
                    <Box>
                      <Typography sx={{ fontSize: '11px', color: colors.textSecondary }}>Chargeability</Typography>
                      <Typography sx={{ fontSize: '13px', fontWeight: 600 }}>
                        {readingDetail.chargeability?.toFixed(4) ?? 'N/A'} mV/V
                      </Typography>
                    </Box>
                  </Box>
                </Paper>

                {/* Level 3 Decay Curve Plot */}
                <DecayCurvePlot decayString={readingDetail.decay_curve} title={`Decay — ${readingDetail.id}`} />

                {/* Actions Panel */}
                <Paper variant="outlined" sx={{ p: 2, borderColor: colors.border, bgcolor: 'background.paper' }}>
                  <Typography sx={{ fontSize: '11px', fontWeight: 600, textTransform: 'uppercase', color: colors.textSecondary, mb: 1.5 }}>
                    Assign Review Verdict
                  </Typography>
                  <Stack spacing={1}>
                    <Button
                      fullWidth
                      variant={readingDetail.qc_review === 'pass' ? 'contained' : 'outlined'}
                      color="success"
                      onClick={() => updateQCStatus(readingDetail.id, 'pass')}
                      startIcon={<CheckCircle size={16} />}
                      sx={{ textTransform: 'none', fontSize: '13px' }}
                    >
                      Mark as Passed
                    </Button>
                    <Button
                      fullWidth
                      variant={readingDetail.qc_review === 'flag' ? 'contained' : 'outlined'}
                      color="warning"
                      onClick={() => updateQCStatus(readingDetail.id, 'flag')}
                      startIcon={<AlertTriangle size={16} />}
                      sx={{ textTransform: 'none', fontSize: '13px' }}
                    >
                      Flag for Inspection
                    </Button>
                    <Button
                      fullWidth
                      variant={readingDetail.qc_review === 'reject' ? 'contained' : 'outlined'}
                      color="error"
                      onClick={() => updateQCStatus(readingDetail.id, 'reject')}
                      startIcon={<XCircle size={16} />}
                      sx={{ textTransform: 'none', fontSize: '13px' }}
                    >
                      Reject Reading
                    </Button>
                  </Stack>
                </Paper>
              </>
            ) : (
              <Box sx={{ py: 6, textAlign: 'center', color: colors.textSecondary }}>
                <Typography variant="body2">
                  Select a geometry from the table to inspect decay curves and assign a QC verdict.
                </Typography>
              </Box>
            )}
          </Box>
        </Box>
      </Box>
    </Box>
  )
}