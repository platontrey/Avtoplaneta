import 'package:flutter/material.dart';
import 'package:flutter_lucide/flutter_lucide.dart';

/// Поле выбора значения из справочника с поиском.
///
/// Остаётся обычным текстовым полем: список — это подсказка, а не ограничение.
/// Справочник не покрывает редкие машины и спецтехнику, поэтому руками ввести
/// своё значение по-прежнему можно, а если справочник не загрузился, поле просто
/// работает как раньше.
class VehiclePickerField extends StatelessWidget {
  const VehiclePickerField({
    super.key,
    required this.controller,
    required this.label,
    required this.options,
    this.hint,
    this.required = false,
    this.onSelected,
  });

  final TextEditingController controller;
  final String label;
  final List<String> options;
  final String? hint;
  final bool required;
  final ValueChanged<String>? onSelected;

  Future<void> _pick(BuildContext context) async {
    if (options.isEmpty) return;
    final selected = await showModalBottomSheet<String>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Theme.of(context).canvasColor,
      builder: (_) => _VehicleOptionsSheet(title: label, options: options),
    );
    if (selected != null) {
      controller.text = selected;
      onSelected?.call(selected);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: TextFormField(
        controller: controller,
        style: const TextStyle(color: Colors.white),
        decoration: InputDecoration(
          labelText: label,
          hintText: hint,
          suffixIcon: options.isEmpty
              ? null
              : IconButton(
                  icon: const Icon(LucideIcons.chevron_down),
                  tooltip: 'Выбрать из справочника',
                  onPressed: () => _pick(context),
                ),
        ),
        onChanged: onSelected,
        validator: required
            ? (value) =>
                (value == null || value.trim().isEmpty) ? 'Обязательное поле' : null
            : null,
      ),
    );
  }
}

class _VehicleOptionsSheet extends StatefulWidget {
  const _VehicleOptionsSheet({required this.title, required this.options});

  final String title;
  final List<String> options;

  @override
  State<_VehicleOptionsSheet> createState() => _VehicleOptionsSheetState();
}

class _VehicleOptionsSheetState extends State<_VehicleOptionsSheet> {
  String _query = '';

  List<String> get _visible {
    final needle = _query.trim().toLowerCase();
    if (needle.isEmpty) return widget.options;
    return widget.options
        .where((option) => option.toLowerCase().contains(needle))
        .toList(growable: false);
  }

  @override
  Widget build(BuildContext context) {
    final visible = _visible;
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.of(context).viewInsets.bottom),
      child: SizedBox(
        height: MediaQuery.of(context).size.height * 0.7,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
              child: TextField(
                autofocus: true,
                style: const TextStyle(color: Colors.white),
                decoration: InputDecoration(
                  labelText: widget.title,
                  hintText: 'Поиск',
                  prefixIcon: const Icon(LucideIcons.search),
                ),
                onChanged: (value) => setState(() => _query = value),
              ),
            ),
            Expanded(
              child: visible.isEmpty
                  ? const Center(
                      child: Text(
                        'Ничего не найдено — закройте список и введите своё значение',
                        textAlign: TextAlign.center,
                      ),
                    )
                  : ListView.builder(
                      itemCount: visible.length,
                      itemBuilder: (_, index) => ListTile(
                        title: Text(visible[index]),
                        onTap: () => Navigator.of(context).pop(visible[index]),
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }
}
