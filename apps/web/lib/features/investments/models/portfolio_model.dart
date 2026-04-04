class PortfolioModel {
  final String id;
  final String userId;
  final String name;
  final String? description;
  final bool isDefault;
  final DateTime createdAt;
  final DateTime updatedAt;

  const PortfolioModel({
    required this.id,
    required this.userId,
    required this.name,
    this.description,
    required this.isDefault,
    required this.createdAt,
    required this.updatedAt,
  });

  factory PortfolioModel.fromJson(Map<String, dynamic> json) {
    return PortfolioModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      isDefault: json['is_default'] as bool? ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'name': name,
        if (description != null) 'description': description,
        'is_default': isDefault,
        'created_at': createdAt.toIso8601String(),
        'updated_at': updatedAt.toIso8601String(),
      };
}
