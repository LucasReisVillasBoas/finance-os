class UserModel {
  final String id;
  final String email;
  final String name;
  final String plan;
  final bool emailVerified;

  const UserModel({
    required this.id,
    required this.email,
    required this.name,
    required this.plan,
    required this.emailVerified,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
        id: json['id'] as String,
        email: json['email'] as String,
        name: json['name'] as String,
        plan: (json['plan'] as String?) ?? 'free',
        emailVerified: (json['email_verified'] as bool?) ?? false,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'email': email,
        'name': name,
        'plan': plan,
        'email_verified': emailVerified,
      };
}
