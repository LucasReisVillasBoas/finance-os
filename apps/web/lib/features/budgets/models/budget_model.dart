import 'package:flutter/material.dart';

class BudgetModel {
  final String id;
  final String? categoryId;
  final double amount;
  final String period;
  final int? month;
  final int? year;
  final double thresholdPct;
  final String? categoryName;
  final String? categoryColor;
  final String? categoryIcon;

  const BudgetModel({
    required this.id,
    this.categoryId,
    required this.amount,
    required this.period,
    this.month,
    this.year,
    required this.thresholdPct,
    this.categoryName,
    this.categoryColor,
    this.categoryIcon,
  });

  factory BudgetModel.fromJson(Map<String, dynamic> json) {
    return BudgetModel(
      id: json['id'] as String,
      categoryId: json['category_id'] as String?,
      amount: (json['amount'] as num).toDouble(),
      period: json['period'] as String,
      month: json['month'] as int?,
      year: json['year'] as int?,
      thresholdPct: (json['threshold_pct'] as num?)?.toDouble() ?? 80.0,
      categoryName: json['category_name'] as String?,
      categoryColor: json['category_color'] as String?,
      categoryIcon: json['category_icon'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      if (categoryId != null) 'category_id': categoryId,
      'amount': amount,
      'period': period,
      if (month != null) 'month': month,
      if (year != null) 'year': year,
      'threshold_pct': thresholdPct,
    };
  }

  BudgetModel copyWith({
    String? id,
    String? categoryId,
    double? amount,
    String? period,
    int? month,
    int? year,
    double? thresholdPct,
    String? categoryName,
    String? categoryColor,
    String? categoryIcon,
  }) {
    return BudgetModel(
      id: id ?? this.id,
      categoryId: categoryId ?? this.categoryId,
      amount: amount ?? this.amount,
      period: period ?? this.period,
      month: month ?? this.month,
      year: year ?? this.year,
      thresholdPct: thresholdPct ?? this.thresholdPct,
      categoryName: categoryName ?? this.categoryName,
      categoryColor: categoryColor ?? this.categoryColor,
      categoryIcon: categoryIcon ?? this.categoryIcon,
    );
  }

  String get periodLabel {
    switch (period) {
      case 'weekly':
        return 'Semanal';
      case 'monthly':
        return 'Mensal';
      case 'yearly':
        return 'Anual';
      default:
        return period;
    }
  }
}

class BudgetProgressModel {
  final String budgetId;
  final String? categoryId;
  final String categoryName;
  final String? categoryColor;
  final String? categoryIcon;
  final double planned;
  final double actual;
  final double percentage;
  final bool isAlert;

  const BudgetProgressModel({
    required this.budgetId,
    this.categoryId,
    required this.categoryName,
    this.categoryColor,
    this.categoryIcon,
    required this.planned,
    required this.actual,
    required this.percentage,
    required this.isAlert,
  });

  factory BudgetProgressModel.fromJson(Map<String, dynamic> json) {
    return BudgetProgressModel(
      budgetId: json['budget_id'] as String,
      categoryId: json['category_id'] as String?,
      categoryName: json['category_name'] as String? ?? 'Geral',
      categoryColor: json['category_color'] as String?,
      categoryIcon: json['category_icon'] as String?,
      planned: (json['planned'] as num).toDouble(),
      actual: (json['actual'] as num).toDouble(),
      percentage: (json['percentage'] as num).toDouble(),
      isAlert: json['is_alert'] as bool? ?? false,
    );
  }

  Color get progressColor {
    if (percentage >= 100) return Colors.red;
    if (percentage >= 70) return Colors.orange;
    return Colors.green;
  }

  double get progressValue => (percentage / 100).clamp(0.0, 1.0);
}
