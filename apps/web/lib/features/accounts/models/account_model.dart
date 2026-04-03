class AccountModel {
  final String id;
  final String userId;
  final String name;
  final String type;
  final String? institution;
  final double balance;
  final double? creditLimit;
  final String? color;
  final String? icon;
  final bool isActive;
  final String createdAt;
  final String updatedAt;

  const AccountModel({
    required this.id,
    required this.userId,
    required this.name,
    required this.type,
    this.institution,
    required this.balance,
    this.creditLimit,
    this.color,
    this.icon,
    required this.isActive,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AccountModel.fromJson(Map<String, dynamic> json) => AccountModel(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        name: json['name'] as String,
        type: json['type'] as String,
        institution: json['institution'] as String?,
        balance: (json['balance'] as num).toDouble(),
        creditLimit: json['credit_limit'] != null
            ? (json['credit_limit'] as num).toDouble()
            : null,
        color: json['color'] as String?,
        icon: json['icon'] as String?,
        isActive: json['is_active'] as bool? ?? true,
        createdAt: json['created_at'] as String? ?? '',
        updatedAt: json['updated_at'] as String? ?? '',
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'name': name,
        'type': type,
        if (institution != null) 'institution': institution,
        'balance': balance,
        if (creditLimit != null) 'credit_limit': creditLimit,
        if (color != null) 'color': color,
        if (icon != null) 'icon': icon,
        'is_active': isActive,
        'created_at': createdAt,
        'updated_at': updatedAt,
      };

  AccountModel copyWith({
    String? id,
    String? userId,
    String? name,
    String? type,
    String? institution,
    double? balance,
    double? creditLimit,
    String? color,
    String? icon,
    bool? isActive,
    String? createdAt,
    String? updatedAt,
  }) =>
      AccountModel(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        name: name ?? this.name,
        type: type ?? this.type,
        institution: institution ?? this.institution,
        balance: balance ?? this.balance,
        creditLimit: creditLimit ?? this.creditLimit,
        color: color ?? this.color,
        icon: icon ?? this.icon,
        isActive: isActive ?? this.isActive,
        createdAt: createdAt ?? this.createdAt,
        updatedAt: updatedAt ?? this.updatedAt,
      );

  /// Human-readable label for the account type.
  String get typeLabel {
    switch (type) {
      case 'checking':
        return 'Conta Corrente';
      case 'savings':
        return 'Poupança';
      case 'credit_card':
        return 'Cartão de Crédito';
      case 'investment':
        return 'Investimento';
      case 'wallet':
        return 'Carteira';
      default:
        return 'Outro';
    }
  }
}
