class TransactionModel {
  final String id;
  final String userId;
  final String accountId;
  final String? categoryId;
  final String type; // income, expense, transfer
  final double amount;
  final String? description;
  final String? notes;
  final DateTime date;
  final List<String> tags;
  final String? transferPairId;
  final bool aiCategorized;
  final double? aiConfidence;
  final String createdAt;
  final String updatedAt;
  // Joined fields
  final String? accountName;
  final String? categoryName;
  final String? categoryColor;
  final String? categoryIcon;

  const TransactionModel({
    required this.id,
    required this.userId,
    required this.accountId,
    this.categoryId,
    required this.type,
    required this.amount,
    this.description,
    this.notes,
    required this.date,
    this.tags = const [],
    this.transferPairId,
    this.aiCategorized = false,
    this.aiConfidence,
    required this.createdAt,
    required this.updatedAt,
    this.accountName,
    this.categoryName,
    this.categoryColor,
    this.categoryIcon,
  });

  factory TransactionModel.fromJson(Map<String, dynamic> json) {
    return TransactionModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      accountId: json['account_id'] as String,
      categoryId: json['category_id'] as String?,
      type: json['type'] as String,
      amount: (json['amount'] as num).toDouble(),
      description: json['description'] as String?,
      notes: json['notes'] as String?,
      date: DateTime.parse(json['date'] as String),
      tags: (json['tags'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          [],
      transferPairId: json['transfer_pair_id'] as String?,
      aiCategorized: json['ai_categorized'] as bool? ?? false,
      aiConfidence: json['ai_confidence'] != null
          ? (json['ai_confidence'] as num).toDouble()
          : null,
      createdAt: json['created_at'] as String? ?? '',
      updatedAt: json['updated_at'] as String? ?? '',
      accountName: json['account_name'] as String?,
      categoryName: json['category_name'] as String?,
      categoryColor: json['category_color'] as String?,
      categoryIcon: json['category_icon'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'account_id': accountId,
        if (categoryId != null) 'category_id': categoryId,
        'type': type,
        'amount': amount,
        if (description != null) 'description': description,
        if (notes != null) 'notes': notes,
        'date': date.toUtc().toIso8601String(),
        'tags': tags,
        if (transferPairId != null) 'transfer_pair_id': transferPairId,
        'ai_categorized': aiCategorized,
        if (aiConfidence != null) 'ai_confidence': aiConfidence,
        'created_at': createdAt,
        'updated_at': updatedAt,
        if (accountName != null) 'account_name': accountName,
        if (categoryName != null) 'category_name': categoryName,
        if (categoryColor != null) 'category_color': categoryColor,
        if (categoryIcon != null) 'category_icon': categoryIcon,
      };

  TransactionModel copyWith({
    String? id,
    String? userId,
    String? accountId,
    String? categoryId,
    String? type,
    double? amount,
    String? description,
    String? notes,
    DateTime? date,
    List<String>? tags,
    String? transferPairId,
    bool? aiCategorized,
    double? aiConfidence,
    String? createdAt,
    String? updatedAt,
    String? accountName,
    String? categoryName,
    String? categoryColor,
    String? categoryIcon,
  }) =>
      TransactionModel(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        accountId: accountId ?? this.accountId,
        categoryId: categoryId ?? this.categoryId,
        type: type ?? this.type,
        amount: amount ?? this.amount,
        description: description ?? this.description,
        notes: notes ?? this.notes,
        date: date ?? this.date,
        tags: tags ?? this.tags,
        transferPairId: transferPairId ?? this.transferPairId,
        aiCategorized: aiCategorized ?? this.aiCategorized,
        aiConfidence: aiConfidence ?? this.aiConfidence,
        createdAt: createdAt ?? this.createdAt,
        updatedAt: updatedAt ?? this.updatedAt,
        accountName: accountName ?? this.accountName,
        categoryName: categoryName ?? this.categoryName,
        categoryColor: categoryColor ?? this.categoryColor,
        categoryIcon: categoryIcon ?? this.categoryIcon,
      );

  bool get isIncome => type == 'income';
  bool get isExpense => type == 'expense';
  bool get isTransfer => type == 'transfer';

  String get typeLabel {
    switch (type) {
      case 'income':
        return 'Receita';
      case 'expense':
        return 'Despesa';
      case 'transfer':
        return 'Transferência';
      default:
        return type;
    }
  }
}
