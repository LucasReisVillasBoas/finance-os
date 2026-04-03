class CategoryModel {
  final String id;
  final String? userId;
  final String name;
  final String type;
  final String? icon;
  final String? color;
  final bool isSystem;
  final String? parentId;
  final bool isActive;
  final String createdAt;

  const CategoryModel({
    required this.id,
    this.userId,
    required this.name,
    required this.type,
    this.icon,
    this.color,
    required this.isSystem,
    this.parentId,
    required this.isActive,
    required this.createdAt,
  });

  factory CategoryModel.fromJson(Map<String, dynamic> json) => CategoryModel(
        id: json['id'] as String,
        userId: json['user_id'] as String?,
        name: json['name'] as String,
        type: json['type'] as String,
        icon: json['icon'] as String?,
        color: json['color'] as String?,
        isSystem: json['is_system'] as bool? ?? false,
        parentId: json['parent_id'] as String?,
        isActive: json['is_active'] as bool? ?? true,
        createdAt: json['created_at'] as String? ?? '',
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        if (userId != null) 'user_id': userId,
        'name': name,
        'type': type,
        if (icon != null) 'icon': icon,
        if (color != null) 'color': color,
        'is_system': isSystem,
        if (parentId != null) 'parent_id': parentId,
        'is_active': isActive,
        'created_at': createdAt,
      };

  CategoryModel copyWith({
    String? id,
    String? userId,
    String? name,
    String? type,
    String? icon,
    String? color,
    bool? isSystem,
    String? parentId,
    bool? isActive,
    String? createdAt,
  }) =>
      CategoryModel(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        name: name ?? this.name,
        type: type ?? this.type,
        icon: icon ?? this.icon,
        color: color ?? this.color,
        isSystem: isSystem ?? this.isSystem,
        parentId: parentId ?? this.parentId,
        isActive: isActive ?? this.isActive,
        createdAt: createdAt ?? this.createdAt,
      );

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
