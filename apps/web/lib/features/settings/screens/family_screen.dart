import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/family_provider.dart';

class FamilyScreen extends ConsumerStatefulWidget {
  const FamilyScreen({super.key});

  @override
  ConsumerState<FamilyScreen> createState() => _FamilyScreenState();
}

class _FamilyScreenState extends ConsumerState<FamilyScreen> {
  final _nameController = TextEditingController();
  final _codeController = TextEditingController();

  @override
  void dispose() {
    _nameController.dispose();
    _codeController.dispose();
    super.dispose();
  }

  Future<void> _createGroup() async {
    if (_nameController.text.trim().isEmpty) return;
    try {
      await ref.read(familyProvider.notifier).createGroup(_nameController.text.trim());
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Grupo criado com sucesso!')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erro: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _joinGroup() async {
    if (_codeController.text.trim().isEmpty) return;
    try {
      await ref.read(familyProvider.notifier).joinGroup(_codeController.text.trim());
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Entrou no grupo com sucesso!')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Código inválido ou já pertence a um grupo')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(familyProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Família')),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : state.group == null
              ? _buildNoGroup()
              : _buildGroup(state),
    );
  }

  Widget _buildNoGroup() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Icon(Icons.group_outlined, size: 64, color: Colors.grey),
          const SizedBox(height: 16),
          const Text(
            'Você não pertence a nenhum grupo familiar.',
            textAlign: TextAlign.center,
            style: TextStyle(fontSize: 16),
          ),
          const SizedBox(height: 32),
          const Text('Criar novo grupo', style: TextStyle(fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          TextField(
            controller: _nameController,
            decoration: const InputDecoration(
              labelText: 'Nome do grupo',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 12),
          ElevatedButton(
            onPressed: _createGroup,
            child: const Text('Criar Grupo'),
          ),
          const SizedBox(height: 32),
          const Divider(),
          const SizedBox(height: 16),
          const Text('Entrar com código', style: TextStyle(fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          TextField(
            controller: _codeController,
            decoration: const InputDecoration(
              labelText: 'Código de convite',
              border: OutlineInputBorder(),
            ),
            textCapitalization: TextCapitalization.characters,
          ),
          const SizedBox(height: 12),
          OutlinedButton(
            onPressed: _joinGroup,
            child: const Text('Entrar no Grupo'),
          ),
        ],
      ),
    );
  }

  Widget _buildGroup(FamilyState state) {
    final group = state.group!;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Card(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  group.name,
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text('Código de convite',
                              style: TextStyle(color: Colors.grey, fontSize: 12)),
                          Text(
                            group.inviteCode,
                            style: const TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.bold,
                              letterSpacing: 3,
                            ),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.copy),
                      onPressed: () {
                        Clipboard.setData(ClipboardData(text: group.inviteCode));
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('Código copiado!')),
                        );
                      },
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 16),
        const Text(
          'Membros',
          style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 8),
        if (state.members.isEmpty)
          const Padding(
            padding: EdgeInsets.all(16),
            child: Text('Nenhum membro além de você.', textAlign: TextAlign.center),
          )
        else
          ...state.members.map(
            (m) => ListTile(
              leading: CircleAvatar(
                child: Text(
                  m.userName.isNotEmpty ? m.userName[0].toUpperCase() : '?',
                ),
              ),
              title: Text(m.userName),
              subtitle: Text(m.userEmail),
              trailing: IconButton(
                icon: const Icon(Icons.remove_circle_outline, color: Colors.red),
                onPressed: () async {
                  final confirm = await showDialog<bool>(
                    context: context,
                    builder: (ctx) => AlertDialog(
                      title: const Text('Remover membro'),
                      content:
                          Text('Remover ${m.userName} do grupo?'),
                      actions: [
                        TextButton(
                          onPressed: () => Navigator.pop(ctx, false),
                          child: const Text('Cancelar'),
                        ),
                        TextButton(
                          onPressed: () => Navigator.pop(ctx, true),
                          child: const Text('Remover',
                              style: TextStyle(color: Colors.red)),
                        ),
                      ],
                    ),
                  );
                  if (confirm == true) {
                    await ref.read(familyProvider.notifier).removeMember(m.id);
                  }
                },
              ),
            ),
          ),
      ],
    );
  }
}
