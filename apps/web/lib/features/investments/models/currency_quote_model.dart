import 'package:flutter/material.dart';

/// A foreign-exchange quote (e.g. USD-BRL, EUR-BRL) returned by
/// GET /api/v1/quotes/currencies.
class CurrencyQuoteModel {
  final String code; // base currency, e.g. "USD"
  final String codein; // quote currency, e.g. "BRL"
  final String name;
  final double bid;
  final double ask;
  final double high;
  final double low;
  final double pctChange;
  final DateTime? updatedAt;

  const CurrencyQuoteModel({
    required this.code,
    required this.codein,
    required this.name,
    required this.bid,
    required this.ask,
    required this.high,
    required this.low,
    required this.pctChange,
    this.updatedAt,
  });

  factory CurrencyQuoteModel.fromJson(Map<String, dynamic> json) {
    return CurrencyQuoteModel(
      code: json['code'] as String? ?? '',
      codein: json['codein'] as String? ?? 'BRL',
      name: json['name'] as String? ?? '',
      bid: (json['bid'] as num?)?.toDouble() ?? 0,
      ask: (json['ask'] as num?)?.toDouble() ?? 0,
      high: (json['high'] as num?)?.toDouble() ?? 0,
      low: (json['low'] as num?)?.toDouble() ?? 0,
      pctChange: (json['pct_change'] as num?)?.toDouble() ?? 0,
      updatedAt: json['updated_at'] != null
          ? DateTime.tryParse(json['updated_at'] as String)
          : null,
    );
  }

  /// "USD/BRL"
  String get pair => '$code/$codein';

  /// A short label like "Dólar" / "Euro" derived from the currency code.
  String get shortLabel {
    switch (code) {
      case 'USD':
        return 'Dólar';
      case 'EUR':
        return 'Euro';
      case 'GBP':
        return 'Libra';
      default:
        return code;
    }
  }

  Color get changeColor => pctChange >= 0
      ? const Color(0xFF22C55E) // green
      : const Color(0xFFEF4444); // red
}
