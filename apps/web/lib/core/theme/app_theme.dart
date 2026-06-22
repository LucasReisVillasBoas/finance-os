import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

// ── Palette ─────────────────────────────────────────────────────────────────

class AppColors {
  AppColors._();

  // Brand
  static const primary = Color(0xFF3B5BDB);
  static const primaryLight = Color(0xFF7395F5);
  static const primaryDark = Color(0xFF2F4AC3);
  static const primaryContainer = Color(0xFFEEF2FF);
  static const onPrimary = Color(0xFFFFFFFF);

  // Accent (investimentos)
  static const violet = Color(0xFF7C3AED);
  static const violetContainer = Color(0xFFF5F0FF);

  // Semantic
  static const income = Color(0xFF059669);
  static const incomeLight = Color(0xFFD1FAE5);
  static const expense = Color(0xFFDC2626);
  static const expenseLight = Color(0xFFFEE2E2);
  static const warning = Color(0xFFD97706);
  static const warningLight = Color(0xFFFEF3C7);

  // Neutrals – light
  static const bgLight = Color(0xFFF8FAFC);
  static const surfaceLight = Color(0xFFFFFFFF);
  static const surfaceElevatedLight = Color(0xFFF1F5F9);
  static const borderLight = Color(0xFFE2E8F0);
  static const textHigh = Color(0xFF0F172A);
  static const textMedium = Color(0xFF475569);
  static const textLow = Color(0xFF94A3B8);

  // Neutrals – dark
  static const bgDark = Color(0xFF0C1525);
  static const surfaceDark = Color(0xFF1A2540);
  static const surfaceElevatedDark = Color(0xFF243050);
  static const borderDark = Color(0xFF2D3F5F);
  static const textHighDark = Color(0xFFF1F5F9);
  static const textMediumDark = Color(0xFF94A3B8);
}

// ── Theme ────────────────────────────────────────────────────────────────────

class AppTheme {
  AppTheme._();

  static TextTheme _textTheme(Color high, Color medium) =>
      GoogleFonts.plusJakartaSansTextTheme().copyWith(
        displayLarge: GoogleFonts.plusJakartaSans(
            fontSize: 36, fontWeight: FontWeight.w700, color: high),
        displayMedium: GoogleFonts.plusJakartaSans(
            fontSize: 28, fontWeight: FontWeight.w700, color: high),
        headlineLarge: GoogleFonts.plusJakartaSans(
            fontSize: 24, fontWeight: FontWeight.w700, color: high),
        headlineMedium: GoogleFonts.plusJakartaSans(
            fontSize: 20, fontWeight: FontWeight.w600, color: high),
        headlineSmall: GoogleFonts.plusJakartaSans(
            fontSize: 18, fontWeight: FontWeight.w600, color: high),
        titleLarge: GoogleFonts.plusJakartaSans(
            fontSize: 16, fontWeight: FontWeight.w600, color: high),
        titleMedium: GoogleFonts.plusJakartaSans(
            fontSize: 14, fontWeight: FontWeight.w600, color: high),
        titleSmall: GoogleFonts.plusJakartaSans(
            fontSize: 13, fontWeight: FontWeight.w500, color: high),
        bodyLarge: GoogleFonts.plusJakartaSans(
            fontSize: 15, fontWeight: FontWeight.w400, color: high),
        bodyMedium: GoogleFonts.plusJakartaSans(
            fontSize: 14, fontWeight: FontWeight.w400, color: medium),
        bodySmall: GoogleFonts.plusJakartaSans(
            fontSize: 12, fontWeight: FontWeight.w400, color: medium),
        labelLarge: GoogleFonts.plusJakartaSans(
            fontSize: 14, fontWeight: FontWeight.w600, color: high),
        labelMedium: GoogleFonts.plusJakartaSans(
            fontSize: 12, fontWeight: FontWeight.w500, color: medium),
        labelSmall: GoogleFonts.plusJakartaSans(
            fontSize: 11, fontWeight: FontWeight.w500, color: medium),
      );

  static ThemeData get light {
    final cs = ColorScheme(
      brightness: Brightness.light,
      primary: AppColors.primary,
      onPrimary: AppColors.onPrimary,
      primaryContainer: AppColors.primaryContainer,
      onPrimaryContainer: AppColors.primaryDark,
      secondary: AppColors.violet,
      onSecondary: Colors.white,
      secondaryContainer: AppColors.violetContainer,
      onSecondaryContainer: AppColors.violet,
      tertiary: AppColors.income,
      onTertiary: Colors.white,
      tertiaryContainer: AppColors.incomeLight,
      onTertiaryContainer: AppColors.income,
      error: AppColors.expense,
      onError: Colors.white,
      errorContainer: AppColors.expenseLight,
      onErrorContainer: AppColors.expense,
      surface: AppColors.surfaceLight,
      onSurface: AppColors.textHigh,
      surfaceContainerHighest: AppColors.surfaceElevatedLight,
      onSurfaceVariant: AppColors.textMedium,
      outline: AppColors.borderLight,
      outlineVariant: AppColors.borderLight,
      shadow: Colors.black,
      scrim: Colors.black,
      inverseSurface: AppColors.textHigh,
      onInverseSurface: AppColors.surfaceLight,
      inversePrimary: AppColors.primaryLight,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: cs,
      scaffoldBackgroundColor: AppColors.bgLight,
      textTheme: _textTheme(AppColors.textHigh, AppColors.textMedium),

      appBarTheme: AppBarTheme(
        backgroundColor: AppColors.surfaceLight,
        foregroundColor: AppColors.textHigh,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        titleTextStyle: GoogleFonts.plusJakartaSans(
          fontSize: 18,
          fontWeight: FontWeight.w700,
          color: AppColors.textHigh,
        ),
        iconTheme: const IconThemeData(color: AppColors.textHigh),
      ),

      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: AppColors.surfaceLight,
        indicatorColor: AppColors.primaryContainer,
        iconTheme: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return const IconThemeData(color: AppColors.primary, size: 22);
          }
          return const IconThemeData(color: AppColors.textMedium, size: 22);
        }),
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return GoogleFonts.plusJakartaSans(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: AppColors.primary,
            );
          }
          return GoogleFonts.plusJakartaSans(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: AppColors.textMedium,
          );
        }),
        elevation: 0,
        shadowColor: Colors.transparent,
        surfaceTintColor: Colors.transparent,
        height: 64,
      ),

      cardTheme: CardThemeData(
        elevation: 0,
        color: AppColors.surfaceLight,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: AppColors.borderLight),
        ),
        margin: EdgeInsets.zero,
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.bgLight,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.borderLight),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.borderLight),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide:
              const BorderSide(color: AppColors.primary, width: 1.5),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.expense),
        ),
        labelStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textMedium),
        hintStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textLow),
      ),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primary,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12)),
          elevation: 0,
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12)),
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12)),
          side: const BorderSide(color: AppColors.borderLight),
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: AppColors.primary,
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),

      floatingActionButtonTheme: const FloatingActionButtonThemeData(
        backgroundColor: AppColors.primary,
        foregroundColor: Colors.white,
        elevation: 2,
        shape: CircleBorder(),
      ),

      dividerTheme: const DividerThemeData(
        color: AppColors.borderLight,
        thickness: 1,
        space: 0,
      ),

      chipTheme: ChipThemeData(
        backgroundColor: AppColors.surfaceElevatedLight,
        selectedColor: AppColors.primaryContainer,
        labelStyle:
            GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w500),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        side: const BorderSide(color: AppColors.borderLight),
      ),

      listTileTheme: ListTileThemeData(
        tileColor: Colors.transparent,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),

      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        backgroundColor: AppColors.textHigh,
        contentTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: Colors.white),
      ),

      dialogTheme: DialogThemeData(
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        backgroundColor: AppColors.surfaceLight,
        elevation: 4,
        titleTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 18,
            fontWeight: FontWeight.w700,
            color: AppColors.textHigh),
        contentTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textMedium),
      ),

      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: AppColors.surfaceLight,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        showDragHandle: true,
      ),
    );
  }

  static ThemeData get dark {
    final cs = ColorScheme(
      brightness: Brightness.dark,
      primary: AppColors.primaryLight,
      onPrimary: AppColors.bgDark,
      primaryContainer: AppColors.primaryDark,
      onPrimaryContainer: AppColors.primaryLight,
      secondary: AppColors.violet,
      onSecondary: Colors.white,
      secondaryContainer: const Color(0xFF2D1B69),
      onSecondaryContainer: const Color(0xFFD8B4FE),
      tertiary: AppColors.income,
      onTertiary: Colors.white,
      tertiaryContainer: const Color(0xFF064E3B),
      onTertiaryContainer: const Color(0xFF6EE7B7),
      error: const Color(0xFFF87171),
      onError: AppColors.bgDark,
      errorContainer: const Color(0xFF7F1D1D),
      onErrorContainer: const Color(0xFFFCA5A5),
      surface: AppColors.surfaceDark,
      onSurface: AppColors.textHighDark,
      surfaceContainerHighest: AppColors.surfaceElevatedDark,
      onSurfaceVariant: AppColors.textMediumDark,
      outline: AppColors.borderDark,
      outlineVariant: AppColors.borderDark,
      shadow: Colors.black,
      scrim: Colors.black,
      inverseSurface: AppColors.textHighDark,
      onInverseSurface: AppColors.surfaceDark,
      inversePrimary: AppColors.primary,
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: cs,
      scaffoldBackgroundColor: AppColors.bgDark,
      textTheme:
          _textTheme(AppColors.textHighDark, AppColors.textMediumDark),

      appBarTheme: AppBarTheme(
        backgroundColor: AppColors.surfaceDark,
        foregroundColor: AppColors.textHighDark,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        titleTextStyle: GoogleFonts.plusJakartaSans(
          fontSize: 18,
          fontWeight: FontWeight.w700,
          color: AppColors.textHighDark,
        ),
      ),

      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: AppColors.surfaceDark,
        indicatorColor: AppColors.primaryDark,
        iconTheme: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return const IconThemeData(
                color: AppColors.primaryLight, size: 22);
          }
          return const IconThemeData(
              color: AppColors.textMediumDark, size: 22);
        }),
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return GoogleFonts.plusJakartaSans(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: AppColors.primaryLight,
            );
          }
          return GoogleFonts.plusJakartaSans(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: AppColors.textMediumDark,
          );
        }),
        elevation: 0,
        shadowColor: Colors.transparent,
        surfaceTintColor: Colors.transparent,
        height: 64,
      ),

      cardTheme: CardThemeData(
        elevation: 0,
        color: AppColors.surfaceDark,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: AppColors.borderDark),
        ),
        margin: EdgeInsets.zero,
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.surfaceElevatedDark,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.borderDark),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: AppColors.borderDark),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide:
              const BorderSide(color: AppColors.primaryLight, width: 1.5),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFF87171)),
        ),
        labelStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textMediumDark),
        hintStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textMediumDark),
      ),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primaryLight,
          foregroundColor: AppColors.bgDark,
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12)),
          elevation: 0,
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12)),
          textStyle: GoogleFonts.plusJakartaSans(
              fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      floatingActionButtonTheme: const FloatingActionButtonThemeData(
        backgroundColor: AppColors.primaryLight,
        foregroundColor: AppColors.bgDark,
        elevation: 2,
        shape: CircleBorder(),
      ),

      dividerTheme: const DividerThemeData(
        color: AppColors.borderDark,
        thickness: 1,
        space: 0,
      ),

      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12)),
        backgroundColor: AppColors.textHighDark,
        contentTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.bgDark),
      ),

      dialogTheme: DialogThemeData(
        shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20)),
        backgroundColor: AppColors.surfaceDark,
        elevation: 4,
        titleTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 18,
            fontWeight: FontWeight.w700,
            color: AppColors.textHighDark),
        contentTextStyle: GoogleFonts.plusJakartaSans(
            fontSize: 14, color: AppColors.textMediumDark),
      ),

      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: AppColors.surfaceDark,
        shape: RoundedRectangleBorder(
          borderRadius:
              BorderRadius.vertical(top: Radius.circular(24)),
        ),
        showDragHandle: true,
      ),
    );
  }
}
