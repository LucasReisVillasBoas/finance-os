class RecurrenceModel {
  final String id;
  final String accountId;
  final String? categoryId;
  final String type;
  final double amount;
  final String? description;
  final String frequency;
  final DateTime startDate;
  final DateTime? endDate;
  final DateTime nextDueDate;
  final bool autoLaunch;
  final bool isActive;
  final String? accountName;
  final String? categoryName;

  const RecurrenceModel({
    required this.id,
    required this.accountId,
    this.categoryId,
    required this.type,
    required this.amount,
    this.description,
    required this.frequency,
    required this.startDate,
    this.endDate,
    required this.nextDueDate,
    required this.autoLaunch,
    required this.isActive,
    this.accountName,
    this.categoryName,
  });

  factory RecurrenceModel.fromJson(Map<String, dynamic> json) {
    return RecurrenceModel(
      id: json['id'] as String,
      accountId: json['account_id'] as String,
      categoryId: json['category_id'] as String?,
      type: json['type'] as String,
      amount: (json['amount'] as num).toDouble(),
      description: json['description'] as String?,
      frequency: json['frequency'] as String,
      startDate: DateTime.parse(json['start_date'] as String),
      endDate: json['end_date'] != null
          ? DateTime.parse(json['end_date'] as String)
          : null,
      nextDueDate: DateTime.parse(json['next_due_date'] as String),
      autoLaunch: json['auto_launch'] as bool? ?? false,
      isActive: json['is_active'] as bool? ?? true,
      accountName: json['account_name'] as String?,
      categoryName: json['category_name'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'account_id': accountId,
      if (categoryId != null) 'category_id': categoryId,
      'type': type,
      'amount': amount,
      if (description != null) 'description': description,
      'frequency': frequency,
      'start_date': startDate.toIso8601String(),
      if (endDate != null) 'end_date': endDate!.toIso8601String(),
      'next_due_date': nextDueDate.toIso8601String(),
      'auto_launch': autoLaunch,
      'is_active': isActive,
    };
  }

  RecurrenceModel copyWith({
    String? id,
    String? accountId,
    String? categoryId,
    String? type,
    double? amount,
    String? description,
    String? frequency,
    DateTime? startDate,
    DateTime? endDate,
    DateTime? nextDueDate,
    bool? autoLaunch,
    bool? isActive,
    String? accountName,
    String? categoryName,
  }) {
    return RecurrenceModel(
      id: id ?? this.id,
      accountId: accountId ?? this.accountId,
      categoryId: categoryId ?? this.categoryId,
      type: type ?? this.type,
      amount: amount ?? this.amount,
      description: description ?? this.description,
      frequency: frequency ?? this.frequency,
      startDate: startDate ?? this.startDate,
      endDate: endDate ?? this.endDate,
      nextDueDate: nextDueDate ?? this.nextDueDate,
      autoLaunch: autoLaunch ?? this.autoLaunch,
      isActive: isActive ?? this.isActive,
      accountName: accountName ?? this.accountName,
      categoryName: categoryName ?? this.categoryName,
    );
  }

  String get frequencyLabel {
    switch (frequency) {
      case 'daily':
        return 'Diário';
      case 'weekly':
        return 'Semanal';
      case 'biweekly':
        return 'Quinzenal';
      case 'monthly':
        return 'Mensal';
      case 'yearly':
        return 'Anual';
      default:
        return frequency;
    }
  }

  String get typeLabel => type == 'income' ? 'Receita' : 'Despesa';
}
