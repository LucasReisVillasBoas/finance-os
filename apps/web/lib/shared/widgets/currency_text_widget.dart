import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

class CurrencyText extends StatelessWidget {
  final double value;
  final TextStyle? style;

  /// When true, applies green for positive and red for negative values.
  final bool colorize;

  /// When true, abbreviates values ≥ 1K as "R$ 1,2K" and ≥ 1M as "R$ 1,2M".
  final bool compact;

  const CurrencyText({
    super.key,
    required this.value,
    this.style,
    this.colorize = false,
    this.compact = false,
  });

  @override
  Widget build(BuildContext context) {
    final String formatted;
    if (compact && value.abs() >= 1000000) {
      formatted = 'R\$ ${(value / 1000000).toStringAsFixed(1)}M';
    } else if (compact && value.abs() >= 1000) {
      formatted = 'R\$ ${(value / 1000).toStringAsFixed(1)}K';
    } else {
      formatted =
          NumberFormat.currency(locale: 'pt_BR', symbol: 'R\$').format(value);
    }

    Color? color;
    if (colorize) {
      color = value >= 0
          ? const Color(0xFF2E7D32) // green.shade800
          : const Color(0xFFC62828); // red.shade800
    }

    final effectiveStyle =
        (style ?? Theme.of(context).textTheme.bodyMedium)?.copyWith(
      color: color ?? style?.color,
    );

    return Text(formatted, style: effectiveStyle);
  }
}
