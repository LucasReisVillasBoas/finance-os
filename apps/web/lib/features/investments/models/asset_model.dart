class AssetModel {
  final String id;
  final String? ticker;
  final String name;
  final String type;
  final String? exchange;
  final String currency;
  final double? currentPrice;
  final DateTime? priceUpdatedAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  const AssetModel({
    required this.id,
    this.ticker,
    required this.name,
    required this.type,
    this.exchange,
    required this.currency,
    this.currentPrice,
    this.priceUpdatedAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory AssetModel.fromJson(Map<String, dynamic> json) {
    return AssetModel(
      id: json['id'] as String,
      ticker: json['ticker'] as String?,
      name: json['name'] as String,
      type: json['type'] as String,
      exchange: json['exchange'] as String?,
      currency: json['currency'] as String? ?? 'BRL',
      currentPrice: (json['current_price'] as num?)?.toDouble(),
      priceUpdatedAt: json['price_updated_at'] != null
          ? DateTime.parse(json['price_updated_at'] as String)
          : null,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        if (ticker != null) 'ticker': ticker,
        'name': name,
        'type': type,
        if (exchange != null) 'exchange': exchange,
        'currency': currency,
        if (currentPrice != null) 'current_price': currentPrice,
      };

  String get displayName => ticker != null ? '$ticker — $name' : name;
}
