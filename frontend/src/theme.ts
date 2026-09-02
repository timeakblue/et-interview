import { createTheme } from '@mui/material/styles'

/**
 * ExploreTech design language, expressed as an MUI theme.
 * See docs/DESIGN_LANGUAGE.md for the source of truth.
 */

export const colors = {
  primary: '#072947',
  primaryHover: '#0a3a5c',
  primaryDark: '#051a2e',
  background: '#ffffff',
  surface: '#f8f9fa',
  border: '#e2e4e7',
  textPrimary: '#1a1a1a',
  textSecondary: '#6b7280',
  textMuted: '#9ca3af',
  success: '#16a34a',
  warning: '#d97706',
  error: '#dc2626',
  info: '#2563eb',
  accent: '#3b82f6',
} as const

const fontFamily = 'Inter, system-ui, -apple-system, sans-serif'

export const theme = createTheme({
  palette: {
    primary: { main: colors.primary, dark: colors.primaryDark, light: colors.primaryHover },
    background: { default: colors.surface, paper: colors.background },
    text: { primary: colors.textPrimary, secondary: colors.textSecondary, disabled: colors.textMuted },
    divider: colors.border,
    success: { main: colors.success },
    warning: { main: colors.warning },
    error: { main: colors.error },
    info: { main: colors.info },
  },
  shape: { borderRadius: 4 },
  typography: {
    fontFamily,
    h1: { fontSize: 18, fontWeight: 600, lineHeight: 1.4 },
    h2: { fontSize: 15, fontWeight: 500, lineHeight: 1.4 },
    body1: { fontSize: 14, fontWeight: 400 },
    body2: { fontSize: 13, fontWeight: 400 },
    caption: { fontSize: 12, fontWeight: 400 },
    // Section headers: 11px / 600 / uppercase
    overline: {
      fontSize: 11,
      fontWeight: 600,
      letterSpacing: '0.05em',
      textTransform: 'uppercase',
      lineHeight: 1.6,
    },
    button: { fontSize: 13, fontWeight: 500, textTransform: 'none' },
  },
  components: {
    MuiButton: {
      defaultProps: { disableElevation: true },
      styleOverrides: {
        root: { transition: 'all 150ms ease-out', padding: '6px 12px' },
      },
    },
    MuiPaper: {
      defaultProps: { elevation: 0 },
      styleOverrides: {
        // Flat surfaces rely on borders rather than shadow for separation.
        root: { backgroundImage: 'none' },
      },
    },
  },
})
