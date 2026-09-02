import { useMemo } from 'react'
import type { EChartsOption } from 'echarts'
import { colors } from '../theme'
import { EChart } from './EChart'

const AXIS_LABEL = {
  color: colors.textSecondary,
  fontSize: 12,
  fontFamily: 'Inter, system-ui, -apple-system, sans-serif',
}

interface DemoBarChartProps {
  values: number[]
}

export function DemoBarChart({ values }: DemoBarChartProps) {
  const option = useMemo<EChartsOption>(
    () => ({
      grid: { top: 16, right: 16, bottom: 32, left: 44 },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        backgroundColor: colors.textPrimary,
        borderWidth: 0,
        padding: [6, 10],
        textStyle: { color: '#ffffff', fontSize: 12 },
      },
      xAxis: {
        type: 'category',
        data: values.map((_, i) => `${i + 1}`),
        axisLabel: AXIS_LABEL,
        axisLine: { lineStyle: { color: colors.border } },
        axisTick: { show: false },
      },
      yAxis: {
        type: 'value',
        axisLabel: AXIS_LABEL,
        splitLine: { lineStyle: { color: colors.border } },
      },
      series: [
        {
          type: 'bar',
          name: 'Value',
          data: values,
          barMaxWidth: 48,
          itemStyle: { color: colors.primary, borderRadius: [4, 4, 0, 0] },
          emphasis: { itemStyle: { color: colors.primaryHover } },
        },
      ],
    }),
    [values],
  )

  return <EChart option={option} />
}
