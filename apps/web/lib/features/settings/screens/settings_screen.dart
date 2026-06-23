import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../auth/providers/auth_provider.dart';
import '../../../shared/providers/theme_provider.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(authProvider).user;
    final themeMode = ref.watch(themeModeProvider);
    final isDark = themeMode == ThemeMode.dark ||
        (themeMode == ThemeMode.system &&
            MediaQuery.platformBrightnessOf(context) == Brightness.dark);

    return Scaffold(
      appBar: AppBar(title: const Text('Configurações')),
      body: ListView(
        children: [
          // Account section
          _SectionHeader(title: 'Conta'),
          ListTile(
            leading: const Icon(Icons.person_outline),
            title: const Text('Perfil'),
            subtitle: Text(user?.name ?? ''),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/profile'),
          ),
          const Divider(indent: 56),
          // Preferences section
          _SectionHeader(title: 'Preferências'),
          SwitchListTile(
            secondary: Icon(
              isDark ? Icons.dark_mode_outlined : Icons.light_mode_outlined,
            ),
            title: const Text('Tema escuro'),
            subtitle: Text(isDark ? 'Ativado' : 'Desativado'),
            value: isDark,
            onChanged: (v) =>
                ref.read(themeModeProvider.notifier).toggleDark(v),
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.notifications_outlined),
            title: const Text('Notificações'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/notifications'),
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.category_outlined),
            title: const Text('Categorias'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/categories'),
          ),
          const Divider(indent: 56),
          // Tools section
          _SectionHeader(title: 'Ferramentas'),
          ListTile(
            leading: const Icon(Icons.upload_file_outlined),
            title: const Text('Importar dados'),
            subtitle: const Text('OFX, CSV'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/import'),
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.chat_bubble_outline),
            title: const Text('WhatsApp Bot'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/whatsapp'),
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.group_outlined),
            title: const Text('Família'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/family'),
          ),
          const Divider(indent: 56),
          // Premium section
          _SectionHeader(title: 'Premium'),
          ListTile(
            leading: const Icon(Icons.star_outline),
            title: const Text('Planos'),
            subtitle: Text('Plano atual: ${(user?.plan ?? 'free').toUpperCase()}'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/plans'),
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.psychology_outlined),
            title: const Text('Assistente IA'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/settings/ai'),
          ),
          const Divider(indent: 56),
          // Info section
          _SectionHeader(title: 'Informações'),
          ListTile(
            leading: const Icon(Icons.info_outline),
            title: const Text('Sobre o FinanceOS'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () {
              showAboutDialog(
                context: context,
                applicationName: 'FinanceOS',
                applicationVersion: '1.0.0',
                applicationLegalese: '© 2025 FinanceOS',
              );
            },
          ),
          const Divider(indent: 56),
          ListTile(
            leading: const Icon(Icons.code),
            title: const Text('Versão'),
            trailing: const Text('1.0.0', style: TextStyle(color: Colors.grey)),
          ),
          const SizedBox(height: 16),
          // Logout button
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: ElevatedButton.icon(
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.red.shade50,
                foregroundColor: Colors.red,
                side: BorderSide(color: Colors.red.shade200),
              ),
              icon: const Icon(Icons.logout),
              label: const Text('Sair'),
              onPressed: () async {
                await ref.read(authProvider.notifier).logout();
                if (context.mounted) context.go('/login');
              },
            ),
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 4),
      child: Text(
        title.toUpperCase(),
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: Theme.of(context).colorScheme.primary,
          letterSpacing: 1.2,
        ),
      ),
    );
  }
}
