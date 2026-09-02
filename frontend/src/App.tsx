import Alert from '@mui/material/Alert'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Skeleton from '@mui/material/Skeleton'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { AlertCircle, BarChart3, RefreshCw } from 'lucide-react'
import { DemoBarChart } from './components/DemoBarChart'
import { useDemoData } from './lib/useDemoData'
import { colors } from './theme'

export default function App() {
  const { status, values, error, reload } = useDemoData()

  return (
    <Box
      sx={{
        height: '100vh',
        width: '100vw',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
      }}
    >
      <Box
        component="header"
        sx={{
          px: 3,
          py: 2.5,
          borderBottom: `1px solid ${colors.border}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 2,
        }}
      >
        <Box>
          <Stack direction="row" spacing={1} sx={{ alignItems: 'center' }}>
            <BarChart3 size={20} strokeWidth={1.5} color={colors.primary} />
            <Typography variant="h1" component="h1">
              Demo
            </Typography>
          </Stack>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            Values served by the backend at <code>/api/demo</code>
          </Typography>
        </Box>
        <Button
          variant="outlined"
          color="inherit"
          onClick={reload}
          disabled={status === 'loading'}
          startIcon={<RefreshCw size={14} strokeWidth={1.5} />}
          sx={{
            borderColor: colors.border,
            color: 'text.primary',
            '&:hover': { borderColor: colors.border, bgcolor: colors.surface },
          }}
        >
          Refresh
        </Button>
      </Box>

      <Box sx={{ flex: 1, overflow: 'auto', bgcolor: colors.surface, p: 3 }}>
        <Box
          sx={{
            height: '100%',
            minHeight: 320,
            bgcolor: 'background.paper',
            border: `1px solid ${colors.border}`,
            borderRadius: '8px',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Typography variant="overline" color="text.secondary" sx={{ px: 2, pt: 1.5 }}>
            Demo values
          </Typography>
          <Box sx={{ flex: 1, minHeight: 0, p: 2, pt: 1 }}>
            <ChartBody status={status} values={values} error={error} onRetry={reload} />
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

interface ChartBodyProps {
  status: 'loading' | 'ready' | 'error'
  values: number[] | null
  error: string | null
  onRetry: () => void
}

function ChartBody({ status, values, error, onRetry }: ChartBodyProps) {
  if (status === 'loading') {
    return <Skeleton variant="rounded" animation="wave" sx={{ width: '100%', height: '100%' }} />
  }

  if (status === 'error') {
    return (
      <Stack
        sx={{
          height: '100%',
          px: 3,
          py: 5,
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box
          sx={{
            width: 64,
            height: 64,
            borderRadius: '12px',
            bgcolor: colors.surface,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            mb: 2,
          }}
        >
          <AlertCircle size={28} strokeWidth={1.5} color={colors.error} />
        </Box>
        <Typography variant="h2">Could not load demo data</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, textAlign: 'center' }}>
          {error}
        </Typography>
        <Button variant="contained" onClick={onRetry} sx={{ mt: 2 }}>
          Try again
        </Button>
      </Stack>
    )
  }

  if (!values?.length) {
    return (
      <Alert severity="info" icon={<AlertCircle size={16} strokeWidth={1.5} />}>
        The backend returned no values.
      </Alert>
    )
  }

  return <DemoBarChart values={values} />
}
