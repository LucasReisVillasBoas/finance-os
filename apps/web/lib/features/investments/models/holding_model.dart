import 'package:flutter/material.dart';

class HoldingModel {
  final String id;
  final String portfolioId;
  final String? assetId;
  final String name;
  final String type;
  final double quantity;
  final double avgPrice;
  final double totalInvested;
  final double currentValue;
  final double unrealizedPnl;
  final double unrealizedPnlPct;
  final double realizedPnl;
  final DateTime createdAt;
  final DateTime updatedAt;
  final String? assetTicker;
  final double? assetCurrentPrice;

  const HoldingModel({
    required this.id,
    required this.portfolioId,
    this.assetId,
    required this.name,
    required this.type,
    required this.quantity,
    required this.avgPrice,
    required this.totalInvested,
    required this.currentValue,
    required this.unrealizedPnl,
    required this.unrealizedPnlPct,
    required this.realizedPnl,
    required this.createdAt,
    required this.updatedAt,
    this.assetTicker,
    this.assetCurrentPrice,
  });

  factory HoldingModel.fromJson(Map<String, dynamic> json) {
    return HoldingModel(
      id: json['id'] as String,
      portfolioId: json['portfolio_id'] as String,
      assetId: json['asset_id'] as String?,
      name: json['name'] as String,
      type: json['type'] as String,
      quantity: (json['quantity'] as num?)?.toDouble() ?? 0.0,
      avgPrice: (json['avg_price'] as num?)?.toDouble() ?? 0.0,
      totalInvested: (json['total_invested'] as num?)?.toDouble() ?? 0.0,
      currentValue: (json['current_value'] as num?)?.toDouble() ?? 0.0,
      unrealizedPnl: (json['unrealized_pnl'] as num?)?.toDouble() ?? 0.0,
      unrealizedPnlPct:
          (json['unrealized_pnl_pct'] as num?)?.toDouble() ?? 0.0,
      realizedPnl: (json['realized_pnl'] as num?)?.toDouble() ?? 0.0,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
      assetTicker: json['asset_ticker'] as String?,
      assetCurrentPrice: (json['asset_current_price'] as num?)?.toDouble(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'portfolio_id': portfolioId,
        if (assetId != null) 'asset_id': assetId,
        'name': name,
        'type': type,
        'quantity': quantity,
        'avg_price': avgPrice,
        'total_invested': totalInvested,
        'current_value': currentValue,
        'unrealized_pnl': unrealizedPnl,
        'unrealized_pnl_pct': unrealizedPnlPct,
        'realized_pnl': realizedPnl,
      };

  Color get pnlColor =>
      unrealizedPnl >= 0 ? const Color(0xFF22C55E) : const Color(0xFFEF4444);

  String get displayTicker => assetTicker ?? name;
}
