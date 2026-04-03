import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/categories_provider.dart';
import '../models/category_model.dart';

class CategoriesScreen extends ConsumerStatefulWidget {
  const CategoriesScreen({super.key});

  @override
  ConsumerState<CategoriesScreen> createState() => _CategoriesScreenState();
}

class _CategoriesScreenState extends ConsumerState<CategoriesScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    Future.microtask(
        () => ref.read(categoriesProvider.notifier).loadCategories());
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(categoriesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Categorias'),
        centerTitle: true,
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'Despesas'),
            Tab(text: 'Receitas'),
          ],
        ),
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.error != null
              ? _ErrorView(
                  message: state.error!,
                  onRetry: () =>
                      ref.read(categoriesProvider.notifier).loadCategories(),
                )
              : TabBarView(
                  controller: _tabController,
                  children: [
                    _CategoryList(
                      categories: state.byType('expense'),
                    ),
                    _CategoryList(
                      categories: state.byType('income'),
                    ),
                  ],
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showCategoryDialog(context, null),
        child: const Icon(Icons.add),
      ),
    );
  }

  void _showCategoryDialog(BuildContext context, CategoryModel? existing) {
    showDialog(
      context: context,
      builder: (ctx) => _CategoryFormDialog(
        existing: existing,
        defaultType: _tabController.index == 0 ? 'expense' : 'income',
      ),
    );
  }
}

class _CategoryList extends ConsumerWidget {
  final List<CategoryModel> categories;

  const _CategoryList({required this.categories});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (categories.isEmpty) {
      return const Center(
        child: Text(
          'Nenhuma categoria encontrada.',
          style: TextStyle(color: Colors.grey),
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: categories.length,
      itemBuilder: (context, index) {
        final cat = categories[index];
        return _CategoryTile(category: cat);
      },
    );
  }
}

class _CategoryTile extends ConsumerWidget {
  final CategoryModel category;

  const _CategoryTile({required this.category});

  Color _parseColor(BuildContext context) {
    if (category.color == null) {
      return Theme.of(context).colorScheme.primary;
    }
    try {
      return Color(
          int.parse(category.color!.replaceFirst('#', '0xFF')));
    } catch (_) {
      return Theme.of(context).colorScheme.primary;
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final color = _parseColor(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color.withValues(alpha: 0.15),
          child: Text(
            category.icon ?? category.name[0].toUpperCase(),
            style: TextStyle(color: color),
          ),
        ),
        title: Text(category.name),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (category.isSystem)
              Container(
                padding: const EdgeInsets.symmetric(
                    horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.blue.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Text(
                  'Sistema',
                  style: TextStyle(fontSize: 11, color: Colors.blue),
                ),
              )
            else ...[
              Container(
                padding: const EdgeInsets.symmetric(
                    horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.green.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Text(
                  'Personalizada',
                  style: TextStyle(fontSize: 11, color: Colors.green),
                ),
              ),
              const SizedBox(width: 4),
              IconButton(
                icon: const Icon(Icons.edit, size: 18),
                onPressed: () => showDialog(
                  context: context,
                  builder: (_) => _CategoryFormDialog(
                    existing: category,
                    defaultType: category.type,
                  ),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.delete, size: 18, color: Colors.red),
                onPressed: () async {
                  final confirmed = await showDialog<bool>(
                    context: context,
                    builder: (ctx) => AlertDialog(
                      title: const Text('Excluir categoria'),
                      content: Text(
                          'Deseja excluir a categoria "${category.name}"?'),
                      actions: [
                        TextButton(
                          onPressed: () => Navigator.pop(ctx, false),
                          child: const Text('Cancelar'),
                        ),
                        TextButton(
                          onPressed: () => Navigator.pop(ctx, true),
                          child: const Text('Excluir',
                              style: TextStyle(color: Colors.red)),
                        ),
                      ],
                    ),
                  );
                  if (confirmed == true) {
                    await ref
                        .read(categoriesProvider.notifier)
                        .deleteCategory(category.id);
                  }
                },
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _CategoryFormDialog extends ConsumerStatefulWidget {
  final CategoryModel? existing;
  final String defaultType;

  const _CategoryFormDialog({
    this.existing,
    required this.defaultType,
  });

  @override
  ConsumerState<_CategoryFormDialog> createState() =>
      _CategoryFormDialogState();
}

class _CategoryFormDialogState extends ConsumerState<_CategoryFormDialog> {
  final _formKey = GlobalKey<FormState>();
  late TextEditingController _nameController;
  late String _selectedType;
  String? _selectedColor;
  String? _selectedIcon;

  static const _categoryTypes = [
    ('income', 'Receita'),
    ('expense', 'Despesa'),
    ('transfer', 'Transferência'),
  ];

  static const _predefinedColors = [
    '#2196F3',
    '#4CAF50',
    '#F44336',
    '#FF9800',
    '#9C27B0',
    '#00BCD4',
    '#795548',
    '#607D8B',
  ];

  static const _predefinedIcons = [
    '🏠', '🍔', '🚗', '💊', '📚', '👕', '🎮', '✈️',
    '💰', '📈', '🎁', '💡', '🏋️', '🐾', '🎵', '💻',
  ];

  @override
  void initState() {
    super.initState();
    _nameController =
        TextEditingController(text: widget.existing?.name ?? '');
    _selectedType = widget.existing?.type ?? widget.defaultType;
    _selectedColor = widget.existing?.color;
    _selectedIcon = widget.existing?.icon;
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final payload = {
      'name': _nameController.text.trim(),
      'type': _selectedType,
      if (_selectedColor != null) 'color': _selectedColor,
      if (_selectedIcon != null) 'icon': _selectedIcon,
    };

    final notifier = ref.read(categoriesProvider.notifier);

    if (widget.existing != null) {
      await notifier.updateCategory(widget.existing!.id, payload);
    } else {
      await notifier.createCategory(payload);
    }

    if (mounted) {
      final state = ref.read(categoriesProvider);
      if (state.error == null) {
        Navigator.pop(context);
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(state.error!)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.existing != null;
    final state = ref.watch(categoriesProvider);

    return AlertDialog(
      title: Text(isEditing ? 'Editar Categoria' : 'Nova Categoria'),
      content: SizedBox(
        width: 400,
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextFormField(
                  controller: _nameController,
                  decoration: const InputDecoration(
                    labelText: 'Nome *',
                    border: OutlineInputBorder(),
                  ),
                  validator: (v) {
                    if (v == null || v.trim().length < 2) {
                      return 'Nome deve ter pelo menos 2 caracteres';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: _selectedType,
                  decoration: const InputDecoration(
                    labelText: 'Tipo *',
                    border: OutlineInputBorder(),
                  ),
                  items: _categoryTypes
                      .map((t) => DropdownMenuItem(
                            value: t.$1,
                            child: Text(t.$2),
                          ))
                      .toList(),
                  onChanged: (v) => setState(() => _selectedType = v!),
                ),
                const SizedBox(height: 12),
                Text(
                  'Ícone',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 4),
                Wrap(
                  spacing: 4,
                  runSpacing: 4,
                  children: _predefinedIcons.map((icon) {
                    final selected = _selectedIcon == icon;
                    return GestureDetector(
                      onTap: () => setState(() => _selectedIcon = icon),
                      child: Container(
                        width: 36,
                        height: 36,
                        decoration: BoxDecoration(
                          border: selected
                              ? Border.all(
                                  color: Theme.of(context).colorScheme.primary,
                                  width: 2)
                              : Border.all(color: Colors.grey.shade300),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Center(child: Text(icon)),
                      ),
                    );
                  }).toList(),
                ),
                const SizedBox(height: 12),
                Text(
                  'Cor',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 4),
                Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: _predefinedColors.map((c) {
                    final color =
                        Color(int.parse(c.replaceFirst('#', '0xFF')));
                    final isSelected = _selectedColor == c;
                    return GestureDetector(
                      onTap: () => setState(() => _selectedColor = c),
                      child: Container(
                        width: 32,
                        height: 32,
                        decoration: BoxDecoration(
                          color: color,
                          shape: BoxShape.circle,
                          border: isSelected
                              ? Border.all(color: Colors.black, width: 2)
                              : null,
                        ),
                        child: isSelected
                            ? const Icon(Icons.check,
                                color: Colors.white, size: 16)
                            : null,
                      ),
                    );
                  }).toList(),
                ),
              ],
            ),
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancelar'),
        ),
        ElevatedButton(
          onPressed: state.isLoading ? null : _submit,
          child: state.isLoading
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : Text(isEditing ? 'Salvar' : 'Criar'),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, size: 48, color: Colors.red),
          const SizedBox(height: 16),
          Text(message, textAlign: TextAlign.center),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: onRetry,
            child: const Text('Tentar novamente'),
          ),
        ],
      ),
    );
  }
}
