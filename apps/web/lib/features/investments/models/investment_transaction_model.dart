class InvestmentTransactionModel {
  final String id;
  final String holdingId;
  final String type;
  final double? quantity;
  final double? price;
  final double fees;
  final double total;
  final DateTime date;
  final String? notes;
  final DateTime createdAt;

  const InvestmentTransactionModel({
    required this.id,
    required this.holdingId,
    required this.type,
    this.quantity,
    this.price,
    required this.fees,
    required this.total,
    required this.date,
    this.notes,
    required this.createdAt,
  });

  factory InvestmentTransactionModel.fromJson(Map<String, dynamic> json) {
    return InvestmentTransactionModel(
      id: json['id'] as String,
      holdingId: json['holding_id'] as String,
      type: json['type'] as String,
      quantity: (json['quantity'] as num?)?.toDouble(),
      price: (json['price'] as num?)?.toDouble(),
      fees: (json['fees'] as num?)?.toDouble() ?? 0.0,
      total: (json['total'] as num?)?.toDouble() ?? 0.0,
      date: DateTime.parse(json['date'] as String),
      notes: json['notes'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'holding_id': holdingId,
        'type': type,
        if (quantity != null) 'quantity': quantity,
        if (price != null) 'price': price,
        'fees': fees,
        'total': total,
        'date': date.toIso8601String(),
        if (notes != null) 'notes': notes,
      };

  String get typeLabel {
    switch (type) {
      case 'buy':
        return 'Compra';
      case 'sell':
        return 'Venda';
      case 'dividend':
        return 'Dividendo';
      case 'split':
        return 'Desdobramento';
      case 'bonus':
        return 'Bonificação';
      default:
        return type;
    }
  }
}
