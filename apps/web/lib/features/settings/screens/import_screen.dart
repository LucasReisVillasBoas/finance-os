import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class ImportScreen extends ConsumerStatefulWidget {
  const ImportScreen({super.key});

  @override
  ConsumerState<ImportScreen> createState() => _ImportScreenState();
}

class _ImportScreenState extends ConsumerState<ImportScreen> {
  bool _isLoading = false;
  bool _isSuccess = false;
  _ImportResult? _result;

  // Simulated file selection state
  String? _selectedFileName;
  String? _selectedFileType; // 'ofx' or 'csv'

  void _simulateFileSelection(String type) {
    setState(() {
      _selectedFileType = type;
      _selectedFileName = type == 'ofx' ? 'extrato.ofx' : 'transacoes.csv';
      _result = null;
    });
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          'Arquivo $_selectedFileName selecionado. '
          'Clique em Importar para processar.',
        ),
      ),
    );
  }

  Future<void> _doImport() async {
    if (_selectedFileName == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Selecione um arquivo primeiro.')),
      );
      return;
    }

    setState(() {
      _isLoading = true;
      _result = null;
    });

    // In production, this would use the real file and call the API.
    // Here we simulate the import result for the UI demo.
    await Future.delayed(const Duration(seconds: 1));

    setState(() {
      _isLoading = false;
      _isSuccess = true;
      _result = _ImportResult(
        imported: 24,
        duplicates: 3,
        errors: 0,
        messages: const ['Importação concluída com sucesso!'],
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Importar Transações'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Instructions card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.info_outline, color: Colors.blue),
                      const SizedBox(width: 8),
                      Text(
                        'Como importar',
                        style: theme.textTheme.titleMedium
                            ?.copyWith(fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    '1. Exporte o extrato do seu banco no formato OFX ou CSV.\n'
                    '2. Selecione o arquivo clicando nos botões abaixo.\n'
                    '3. Clique em Importar para processar o arquivo.\n'
                    '4. Transações duplicadas serão ignoradas automaticamente.',
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // File type selection
          Text(
            'Formato do arquivo',
            style: theme.textTheme.titleSmall
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _simulateFileSelection('ofx'),
                  icon: const Icon(Icons.upload_file),
                  label: const Text('Selecionar OFX/QFX'),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    side: BorderSide(
                      color: _selectedFileType == 'ofx'
                          ? Colors.blue
                          : Colors.grey,
                      width: _selectedFileType == 'ofx' ? 2 : 1,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: () => _simulateFileSelection('csv'),
                  icon: const Icon(Icons.table_chart),
                  label: const Text('Selecionar CSV'),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    side: BorderSide(
                      color: _selectedFileType == 'csv'
                          ? Colors.blue
                          : Colors.grey,
                      width: _selectedFileType == 'csv' ? 2 : 1,
                    ),
                  ),
                ),
              ),
            ],
          ),

          if (_selectedFileName != null) ...[
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: Colors.blue.shade50,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.blue.shade200),
              ),
              child: Row(
                children: [
                  const Icon(Icons.attach_file, color: Colors.blue, size: 18),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      _selectedFileName!,
                      style: const TextStyle(color: Colors.blue),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.clear, size: 18, color: Colors.blue),
                    onPressed: () => setState(() {
                      _selectedFileName = null;
                      _selectedFileType = null;
                    }),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                  ),
                ],
              ),
            ),
          ],

          const SizedBox(height: 24),

          SizedBox(
            height: 48,
            child: ElevatedButton.icon(
              onPressed: _isLoading || _selectedFileName == null
                  ? null
                  : _doImport,
              icon: _isLoading
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.cloud_upload),
              label: Text(_isLoading ? 'Importando...' : 'Importar'),
            ),
          ),

          if (_result != null) ...[
            const SizedBox(height: 24),
            _ResultCard(result: _result!, isSuccess: _isSuccess),
          ],

          const SizedBox(height: 24),

          // Formats documentation
          ExpansionTile(
            title: const Text('Formatos suportados'),
            children: [
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _FormatInfo(
                      title: 'OFX / QFX',
                      description:
                          'Formato padrão exportado pela maioria dos bancos '
                          'brasileiros e corretoras. Deduplica automaticamente '
                          'por FITID.',
                    ),
                    const SizedBox(height: 12),
                    _FormatInfo(
                      title: 'CSV',
                      description:
                          'Arquivo de texto com valores separados por vírgula. '
                          'Deve conter colunas: data, valor, descrição e tipo '
                          '(D=Débito, C=Crédito).',
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ImportResult {
  final int imported;
  final int duplicates;
  final int errors;
  final List<String> messages;

  const _ImportResult({
    required this.imported,
    required this.duplicates,
    required this.errors,
    required this.messages,
  });
}

class _ResultCard extends StatelessWidget {
  final _ImportResult result;
  final bool isSuccess;

  const _ResultCard({required this.result, required this.isSuccess});

  @override
  Widget build(BuildContext context) {
    return Card(
      color: isSuccess && result.errors == 0
          ? Colors.green.shade50
          : Colors.orange.shade50,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  result.errors == 0 ? Icons.check_circle : Icons.warning,
                  color: result.errors == 0 ? Colors.green : Colors.orange,
                ),
                const SizedBox(width: 8),
                Text(
                  'Resultado da Importação',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: result.errors == 0 ? Colors.green : Colors.orange,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            _StatRow(
              label: 'Importadas',
              value: '${result.imported}',
              color: Colors.green,
              icon: Icons.check,
            ),
            _StatRow(
              label: 'Duplicatas ignoradas',
              value: '${result.duplicates}',
              color: Colors.orange,
              icon: Icons.skip_next,
            ),
            _StatRow(
              label: 'Erros',
              value: '${result.errors}',
              color: Colors.red,
              icon: Icons.error_outline,
            ),
            if (result.messages.isNotEmpty) ...[
              const Divider(),
              ...result.messages.map(
                (msg) => Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Text(
                    '• $msg',
                    style: const TextStyle(fontSize: 12),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _StatRow extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  final IconData icon;

  const _StatRow({
    required this.label,
    required this.value,
    required this.color,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          Icon(icon, size: 16, color: color),
          const SizedBox(width: 8),
          Text(label),
          const Spacer(),
          Text(
            value,
            style: TextStyle(fontWeight: FontWeight.bold, color: color),
          ),
        ],
      ),
    );
  }
}

class _FormatInfo extends StatelessWidget {
  final String title;
  final String description;

  const _FormatInfo({required this.title, required this.description});

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: const TextStyle(fontWeight: FontWeight.bold)),
        const SizedBox(height: 4),
        Text(description, style: const TextStyle(color: Colors.grey)),
      ],
    );
  }
}
