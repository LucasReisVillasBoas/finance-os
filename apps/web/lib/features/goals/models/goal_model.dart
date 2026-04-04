class GoalModel {
  final String id;
  final String userId;
  final String name;
  final double targetAmount;
  final double currentAmount;
  final DateTime? targetDate;
  final double? monthlyContribution;
  final String? icon;
  final String? color;
  final bool isAchieved;
  final DateTime createdAt;
  final DateTime updatedAt;

  const GoalModel({
    required this.id,
    required this.userId,
    required this.name,
    required this.targetAmount,
    required this.currentAmount,
    this.targetDate,
    this.monthlyContribution,
    this.icon,
    this.color,
    required this.isAchieved,
    required this.createdAt,
    required this.updatedAt,
  });

  double get progressPct =>
      targetAmount > 0 ? (currentAmount / targetAmount).clamp(0.0, 1.0) : 0;

  double get remainingAmount =>
      (targetAmount - currentAmount).clamp(0.0, double.infinity);

  factory GoalModel.fromJson(Map<String, dynamic> json) {
    return GoalModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      name: json['name'] as String,
      targetAmount: (json['target_amount'] as num).toDouble(),
      currentAmount: (json['current_amount'] as num? ?? 0).toDouble(),
      targetDate: json['target_date'] != null
          ? DateTime.tryParse(json['target_date'] as String)
          : null,
      monthlyContribution: json['monthly_contribution'] != null
          ? (json['monthly_contribution'] as num).toDouble()
          : null,
      icon: json['icon'] as String?,
      color: json['color'] as String?,
      isAchieved: json['is_achieved'] as bool? ?? false,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'name': name,
      'target_amount': targetAmount,
      'current_amount': currentAmount,
      if (targetDate != null) 'target_date': targetDate!.toIso8601String(),
      if (monthlyContribution != null)
        'monthly_contribution': monthlyContribution,
      if (icon != null) 'icon': icon,
      if (color != null) 'color': color,
      'is_achieved': isAchieved,
    };
  }
}

class GoalProjectionModel {
  final String goalId;
  final String name;
  final double targetAmount;
  final double currentAmount;
  final double remainingAmount;
  final double progressPct;
  final int? monthsToGoal;
  final DateTime? estimatedDate;

  const GoalProjectionModel({
    required this.goalId,
    required this.name,
    required this.targetAmount,
    required this.currentAmount,
    required this.remainingAmount,
    required this.progressPct,
    this.monthsToGoal,
    this.estimatedDate,
  });

  factory GoalProjectionModel.fromJson(Map<String, dynamic> json) {
    return GoalProjectionModel(
      goalId: json['goal_id'] as String,
      name: json['name'] as String,
      targetAmount: (json['target_amount'] as num).toDouble(),
      currentAmount: (json['current_amount'] as num? ?? 0).toDouble(),
      remainingAmount: (json['remaining_amount'] as num? ?? 0).toDouble(),
      progressPct: (json['progress_pct'] as num? ?? 0).toDouble(),
      monthsToGoal: json['months_to_goal'] as int?,
      estimatedDate: json['estimated_date'] != null
          ? DateTime.tryParse(json['estimated_date'] as String)
          : null,
    );
  }
}
