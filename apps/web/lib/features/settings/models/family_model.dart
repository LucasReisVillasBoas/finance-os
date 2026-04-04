class FamilyGroup {
  final String id;
  final String name;
  final String ownerId;
  final String inviteCode;
  final DateTime createdAt;

  const FamilyGroup({
    required this.id,
    required this.name,
    required this.ownerId,
    required this.inviteCode,
    required this.createdAt,
  });

  factory FamilyGroup.fromJson(Map<String, dynamic> json) {
    return FamilyGroup(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      ownerId: json['owner_id'] as String? ?? '',
      inviteCode: json['invite_code'] as String? ?? '',
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }
}

class FamilyMember {
  final String id;
  final String groupId;
  final String userId;
  final String userName;
  final String userEmail;
  final DateTime joinedAt;

  const FamilyMember({
    required this.id,
    required this.groupId,
    required this.userId,
    required this.userName,
    required this.userEmail,
    required this.joinedAt,
  });

  factory FamilyMember.fromJson(Map<String, dynamic> json) {
    return FamilyMember(
      id: json['id'] as String,
      groupId: json['group_id'] as String? ?? '',
      userId: json['user_id'] as String? ?? '',
      userName: json['user_name'] as String? ?? '',
      userEmail: json['user_email'] as String? ?? '',
      joinedAt: json['joined_at'] != null
          ? DateTime.parse(json['joined_at'] as String)
          : DateTime.now(),
    );
  }
}
