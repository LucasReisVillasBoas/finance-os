import 'package:flutter/material.dart';

class NotificationModel {
  final String id;
  final String type;
  final String title;
  final String? message;
  final bool isRead;
  final DateTime createdAt;

  const NotificationModel({
    required this.id,
    required this.type,
    required this.title,
    this.message,
    required this.isRead,
    required this.createdAt,
  });

  IconData get icon {
    switch (type) {
      case 'budget_alert':
        return Icons.warning_amber;
      case 'goal_deadline':
        return Icons.flag;
      case 'recurrence_due':
        return Icons.repeat;
      case 'weekly_summary':
        return Icons.summarize;
      case 'monthly_report':
        return Icons.calendar_month;
      default:
        return Icons.notifications;
    }
  }

  Color get iconColor {
    switch (type) {
      case 'budget_alert':
        return Colors.orange;
      case 'goal_deadline':
        return Colors.blue;
      case 'recurrence_due':
        return Colors.purple;
      case 'weekly_summary':
      case 'monthly_report':
        return Colors.teal;
      default:
        return Colors.grey;
    }
  }

  factory NotificationModel.fromJson(Map<String, dynamic> json) {
    return NotificationModel(
      id: json['id'] as String,
      type: json['type'] as String? ?? 'info',
      title: json['title'] as String? ?? '',
      message: json['message'] as String?,
      isRead: json['is_read'] as bool? ?? false,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type,
        'title': title,
        'message': message,
        'is_read': isRead,
        'created_at': createdAt.toUtc().toIso8601String(),
      };

  NotificationModel copyWith({bool? isRead}) {
    return NotificationModel(
      id: id,
      type: type,
      title: title,
      message: message,
      isRead: isRead ?? this.isRead,
      createdAt: createdAt,
    );
  }
}
