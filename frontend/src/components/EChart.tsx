import { useEffect, useRef } from 'react'
import type { EChartsOption } from 'echarts'
import { BarChart } from 'echarts/charts'
import { GridComponent, TitleComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import Box from '@mui/material/Box'
import type { SxProps, Theme } from '@mui/material/styles'

// Tree-shaken registration: only the pieces the app actually draws.
echarts.use([BarChart, GridComponent, TitleComponent, TooltipComponent, CanvasRenderer])

interface EChartProps {
  option: EChartsOption
  /** Clears the previous option instead of merging into it. */
  notMerge?: boolean
  sx?: SxProps<Theme>
}

/**
 * Thin React wrapper around an ECharts instance: owns init/dispose and keeps
 * the chart sized to its container.
 */
export function EChart({ option, notMerge = true, sx }: EChartProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const chartRef = useRef<echarts.ECharts | null>(null)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const chart = echarts.init(container, undefined, { renderer: 'canvas' })
    chartRef.current = chart

    const observer = new ResizeObserver(() => chart.resize())
    observer.observe(container)

    return () => {
      observer.disconnect()
      chart.dispose()
      chartRef.current = null
    }
  }, [])

  useEffect(() => {
    chartRef.current?.setOption(option, { notMerge })
  }, [option, notMerge])

  return <Box ref={containerRef} sx={{ width: '100%', height: '100%', ...sx }} />
}
