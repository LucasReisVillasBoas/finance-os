class CustomAssetModel {
  final String id;
  final String userId;
  final String name;
  final String type;
  final double currentValue;
  final double? purchaseValue;
  final DateTime? purchaseDate;
  final double monthlyIncome;
  final String? description;
  final bool isActive;
  final DateTime createdAt;
  final DateTime updatedAt;

  const CustomAssetModel({
    required this.id,
    required this.userId,
    required this.name,
    required this.type,
    required this.currentValue,
    this.purchaseValue,
    this.purchaseDate,
    required this.monthlyIncome,
    this.description,
    required this.isActive,
    required this.createdAt,
    required this.updatedAt,
  });

  factory CustomAssetModel.fromJson(Map<String, dynamic> json) {
    return CustomAssetModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      type: json['type'] as String,
      currentValue: (json['current_value'] as num?)?.toDouble() ?? 0.0,
      purchaseValue: (json['purchase_value'] as num?)?.toDouble(),
      purchaseDate: json['purchase_date'] != null
          ? DateTime.parse(json['purchase_date'] as String)
          : null,
      monthlyIncome: (json['monthly_income'] as num?)?.toDouble() ?? 0.0,
      description: json['description'] as String?,
      isActive: json['is_active'] as bool? ?? true,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'name': name,
        'type': type,
        'current_value': currentValue,
        if (purchaseValue != null) 'purchase_value': purchaseValue,
        if (purchaseDate != null) 'purchase_date': purchaseDate!.toIso8601String(),
        'monthly_income': monthlyIncome,
        if (description != null) 'description': description,
        'is_active': isActive,
      };

  double get unrealizedPnl =>
      purchaseValue != null ? currentValue - purchaseValue! : 0.0;
}
