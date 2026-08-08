package com.jstauff.feats.android.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Shapes
import androidx.compose.material3.Typography
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

private val FeatsLightColors = lightColorScheme(
    primary = FeatsBlue,
    onPrimary = Color.White,
    primaryContainer = FeatsBlueTint,
    onPrimaryContainer = FeatsBlueDark,
    secondary = Color(0xFF3C9D7A),
    onSecondary = Color.White,
    background = FeatsBackground,
    onBackground = FeatsText,
    surface = FeatsSurface,
    onSurface = FeatsText,
    surfaceVariant = Color(0xFFF0F3F8),
    onSurfaceVariant = FeatsMuted,
    error = FeatsError,
    onError = Color.White
)

private val FeatsDarkColors = darkColorScheme(
    primary = FeatsBlueLight,
    onPrimary = Color(0xFF0B1B2E),
    primaryContainer = FeatsBlueDeepTint,
    onPrimaryContainer = FeatsBlueLight,
    secondary = Color(0xFF5FBF9B),
    onSecondary = Color(0xFF0B241C),
    background = FeatsBackgroundDark,
    onBackground = FeatsTextDark,
    surface = FeatsSurfaceDark,
    onSurface = FeatsTextDark,
    surfaceVariant = FeatsSurfaceVariantDark,
    onSurfaceVariant = FeatsMutedDark,
    error = FeatsErrorDark,
    onError = Color(0xFF2B0A06)
)

private val FeatsTypography = Typography(
    headlineLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 30.sp),
    headlineMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Bold, fontSize = 24.sp),
    titleLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 21.sp),
    titleMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 18.sp),
    bodyLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 16.sp),
    bodyMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 14.sp),
    labelLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Medium, fontSize = 13.sp)
)

private val FeatsShapes = Shapes(
    extraSmall = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
    small = androidx.compose.foundation.shape.RoundedCornerShape(12.dp),
    medium = androidx.compose.foundation.shape.RoundedCornerShape(16.dp),
    large = androidx.compose.foundation.shape.RoundedCornerShape(22.dp)
)

/**
 * Brand palette in both light and dark. Material You dynamic color is deliberately
 * not used — Feats keeps its own brand colours across platforms.
 */
@Composable
fun FeatsTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    MaterialTheme(
        colorScheme = if (darkTheme) FeatsDarkColors else FeatsLightColors,
        typography = FeatsTypography,
        shapes = FeatsShapes,
        content = content
    )
}
