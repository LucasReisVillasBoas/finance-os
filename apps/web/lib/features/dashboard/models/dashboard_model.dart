class DashboardOverview {
  final double netBalance;
  final double totalIncome;
  final double totalExpense;
  final double totalPatrimony;
  final double investmentValue;
  final double customAssetValue;
  final double totalNetWorth;
  final double investmentCapacity;
  final double investmentCapacityPct;
  final List<CategorySummaryModel> topCategories;
  final List<BudgetAlertModel> alertBudgets;
  final List<RecentTransactionModel> recentTransactions;

  const DashboardOverview({
    required this.netBalance,
    required this.totalIncome,
    required this.totalExpense,
    required this.totalPatrimony,
    required this.investmentValue,
    required this.customAssetValue,
    required this.totalNetWorth,
    required this.investmentCapacity,
    required this.investmentCapacityPct,
    required this.topCategories,
    required this.alertBudgets,
    required this.recentTransactions,
  });

  factory DashboardOverview.fromJson(Map<String, dynamic> json) {
    return DashboardOverview(
      netBalance: (json['net_balance'] as num?)?.toDouble() ?? 0.0,
      totalIncome: (json['total_income'] as num?)?.toDouble() ?? 0.0,
      totalExpense: (json['total_expense'] as num?)?.toDouble() ?? 0.0,
      totalPatrimony: (json['total_patrimony'] as num?)?.toDouble() ?? 0.0,
      investmentValue: (json['investment_value'] as num?)?.toDouble() ?? 0.0,
      customAssetValue: (json['custom_asset_value'] as num?)?.toDouble() ?? 0.0,
      totalNetWorth: (json['total_net_worth'] as num?)?.toDouble() ?? 0.0,
      investmentCapacity: (json['investment_capacity'] as num?)?.toDouble() ?? 0.0,
      investmentCapacityPct: (json['investment_capacity_pct'] as num?)?.toDouble() ?? 0.0,
      topCategories: (json['top_categories'] as List<dynamic>? ?? [])
          .map((e) => CategorySummaryModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      alertBudgets: (json['alert_budgets'] as List<dynamic>? ?? [])
          .map((e) => BudgetAlertModel.fromJson(e as Map<String, dynamic>))
          .toList(),
      recentTransactions: (json['recent_transactions'] as List<dynamic>? ?? [])
          .map((e) => RecentTransactionModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

class PatrimonySnapshotModel {
  final int month;
  final int year;
  final String label;
  final double bankSavings;
  final double investedTotal;
  final double totalNetWorth;

  const PatrimonySnapshotModel({
    required this.month,
    required this.year,
    required this.label,
    required this.bankSavings,
    required this.investedTotal,
    required this.totalNetWorth,
  });

  factory PatrimonySnapshotModel.fromJson(Map<String, dynamic> json) {
    return PatrimonySnapshotModel(
      month: json['month'] as int,
      year: json['year'] as int,
      label: json['label'] as String? ?? '',
      bankSavings: (json['bank_savings'] as num?)?.toDouble() ?? 0.0,
      investedTotal: (json['invested_total'] as num?)?.toDouble() ?? 0.0,
      totalNetWorth: (json['total_net_worth'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class CategorySummaryModel {
  final String? categoryId;
  final String categoryName;
  final double total;
  final int count;
  final String? color;

  const CategorySummaryModel({
    this.categoryId,
    required this.categoryName,
    required this.total,
    required this.count,
    this.color,
  });

  factory CategorySummaryModel.fromJson(Map<String, dynamic> json) {
    return CategorySummaryModel(
      categoryId: json['category_id'] as String?,
      categoryName: json['category_name'] as String? ?? 'Sem categoria',
      total: (json['total'] as num?)?.toDouble() ?? 0.0,
      count: json['count'] as int? ?? 0,
      color: json['color'] as String?,
    );
  }
}

class BudgetAlertModel {
  final String budgetId;
  final String? categoryId;
  final String categoryName;
  final double planned;
  final double actual;
  final double percentage;
  final bool isAlert;

  const BudgetAlertModel({
    required this.budgetId,
    this.categoryId,
    required this.categoryName,
    required this.planned,
    required this.actual,
    required this.percentage,
    required this.isAlert,
  });

  factory BudgetAlertModel.fromJson(Map<String, dynamic> json) {
    return BudgetAlertModel(
      budgetId: json['budget_id'] as String,
      categoryId: json['category_id'] as String?,
      categoryName: json['category_name'] as String? ?? 'Geral',
      planned: (json['planned'] as num?)?.toDouble() ?? 0.0,
      actual: (json['actual'] as num?)?.toDouble() ?? 0.0,
      percentage: (json['percentage'] as num?)?.toDouble() ?? 0.0,
      isAlert: json['is_alert'] as bool? ?? false,
    );
  }

  double get progressValue => (percentage / 100).clamp(0.0, 1.0);
}

class RecentTransactionModel {
  final String id;
  final String accountId;
  final String? categoryId;
  final String type;
  final double amount;
  final String? description;
  final String date;
  final String? accountName;
  final String? categoryName;
  final String? categoryColor;
  final String? categoryIcon;

  const RecentTransactionModel({
    required this.id,
    required this.accountId,
    this.categoryId,
    required this.type,
    required this.amount,
    this.description,
    required this.date,
    this.accountName,
    this.categoryName,
    this.categoryColor,
    this.categoryIcon,
  });

  factory RecentTransactionModel.fromJson(Map<String, dynamic> json) {
    return RecentTransactionModel(
      id: json['id'] as String,
      accountId: json['account_id'] as String,
      categoryId: json['category_id'] as String?,
      type: json['type'] as String,
      amount: (json['amount'] as num).toDouble(),
      description: json['description'] as String?,
      date: json['date'] as String? ?? '',
      accountName: json['account_name'] as String?,
      categoryName: json['category_name'] as String?,
      categoryColor: json['category_color'] as String?,
      categoryIcon: json['category_icon'] as String?,
    );
  }

  bool get isExpense => type == 'expense';
  bool get isIncome => type == 'income';
  bool get isTransfer => type == 'transfer';
}

class MonthlyCashflowModel {
  final int month;
  final int year;
  final String label;
  final double income;
  final double expense;
  final double balance;

  const MonthlyCashflowModel({
    required this.month,
    required this.year,
    required this.label,
    required this.income,
    required this.expense,
    required this.balance,
  });

  factory MonthlyCashflowModel.fromJson(Map<String, dynamic> json) {
    return MonthlyCashflowModel(
      month: json['month'] as int,
      year: json['year'] as int,
      label: json['label'] as String? ?? '',
      income: (json['income'] as num?)?.toDouble() ?? 0.0,
      expense: (json['expense'] as num?)?.toDouble() ?? 0.0,
      balance: (json['balance'] as num?)?.toDouble() ?? 0.0,
    );
  }
}
