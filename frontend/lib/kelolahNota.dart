import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class NotaSettingsPage extends StatefulWidget {
  final int outletId;
  final String outletName;
  final String outletAddress;
  final String outletPhone;

  const NotaSettingsPage({
    super.key,
    required this.outletId,
    required this.outletName,
    required this.outletAddress,
    required this.outletPhone,
  });

  @override
  State<NotaSettingsPage> createState() => _NotaSettingsPageState();
}

class _NotaSettingsPageState extends State<NotaSettingsPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  String? _accessToken;
  bool _isLoading = false;
  bool _isSaving = false;

  final TextEditingController _footerNoteController = TextEditingController();
  final TextEditingController _whatsappNoteController = TextEditingController();

  bool _showLogo = false;
  bool _showQRCode = true;
  bool _showBusinessName = true;
  bool _showDescription = true;
  bool _showFooterNote = true;
  bool _showWhatsappFooter = true;

  String _printerSize = '58';
  String _printerType = 'A';

  static const Color primaryColor = Color(0xFFEF5350);
  static const Color backgroundColor = Color(0xFFF5F7FA);
  static const Color cardColor = Colors.white;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _footerNoteController.dispose();
    _whatsappNoteController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() => _accessToken = prefs.getString('access_token'));
    await _fetchNotaSettings();
  }

  Future<void> _fetchNotaSettings() async {
    setState(() => _isLoading = true);
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/nota-settings/outlet/${widget.outletId}'),
        headers: {
          'Authorization': 'Bearer $_accessToken',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final settings = data['data'];

        setState(() {
          _footerNoteController.text = settings['footer_note'] ?? '';
          _whatsappNoteController.text = settings['whatsapp_note'] ?? '';

          _showLogo = settings['show_logo'] ?? false;
          _showQRCode = settings['show_qr_code'] ?? true;
          _showBusinessName = settings['show_business_name'] ?? true;
          _showDescription = settings['show_description'] ?? true;
          _showFooterNote = settings['show_footer_note'] ?? true;
          _showWhatsappFooter = settings['show_whatsapp_footer'] ?? true;

          _printerSize = settings['printer_size']?.toString() ?? '58';
          _printerType = settings['printer_type'] ?? 'A';
        });
      } else if (response.statusCode == 404) {
        print('Nota settings belum ada, menggunakan default values');
      } else {
        _showSnackBar('Gagal memuat data: ${response.statusCode}', false);
      }
    } catch (e) {
      print('Error fetching nota settings: $e');
      _showSnackBar('Error: $e', false);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _saveNotaSettings() async {
    setState(() => _isSaving = true);
    try {
      final data = {
        'business_name': widget.outletName,
        'address': widget.outletAddress,
        'phone': widget.outletPhone,
        'footer_note': _footerNoteController.text,
        'whatsapp_note': _whatsappNoteController.text,
        'show_logo': _showLogo,
        'show_qr_code': _showQRCode,
        'show_business_name': _showBusinessName,
        'show_description': _showDescription,
        'show_footer_note': _showFooterNote,
        'show_whatsapp_footer': _showWhatsappFooter,
        'printer_size': int.parse(_printerSize),
        'printer_type': _printerType,
      };

      final response = await http.post(
        Uri.parse('$_baseUrl/nota-settings/outlet/${widget.outletId}'),
        headers: {
          'Authorization': 'Bearer $_accessToken',
          'Content-Type': 'application/json',
        },
        body: json.encode(data),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        _showSnackBar('Pengaturan nota berhasil disimpan', true);

        await Future.delayed(const Duration(milliseconds: 1500));
        if (!mounted) return;
        Navigator.pop(context, true);
      } else {
        try {
          final error = json.decode(response.body);
          _showSnackBar(error['message'] ?? 'Gagal menyimpan', false);
        } catch (e) {
          _showSnackBar('Gagal menyimpan pengaturan', false);
        }
      }
    } catch (e) {
      print('Error saving nota settings: $e');
      _showSnackBar('Error: $e', false);
    } finally {
      setState(() => _isSaving = false);
    }
  }

  void _showSnackBar(String message, bool isSuccess) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(
              isSuccess ? Icons.check_circle : Icons.error_outline,
              color: Colors.white,
            ),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: isSuccess ? Colors.green : Colors.red,
        behavior: SnackBarBehavior.floating,
        duration: Duration(seconds: isSuccess ? 2 : 3),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  Widget _buildSectionHeader(String title, {String? subtitle, IconData? icon}) {
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 20, 16, 12),
      child: Row(
        children: [
          if (icon != null) ...[
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: primaryColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(icon, size: 20, color: primaryColor),
            ),
            const SizedBox(width: 12),
          ],
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.black87,
                  ),
                ),
                if (subtitle != null) ...[
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCard({required Widget child}) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      decoration: BoxDecoration(
        color: cardColor,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: child,
    );
  }

  // Widget untuk menampilkan info read-only
  Widget _buildReadOnlyField({
    required String label,
    required String value,
    required IconData icon,
  }) {
    return _buildCard(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 18, color: primaryColor),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: Colors.black54,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.grey[100],
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey[300]!),
              ),
              child: Text(
                value,
                style: const TextStyle(
                  fontSize: 14,
                  color: Colors.black87,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.info_outline, size: 14, color: Colors.grey[500]),
                const SizedBox(width: 4),
                Text(
                  'Data diambil dari profil outlet',
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.grey[500],
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    int maxLines = 1,
    TextInputType? keyboardType,
    IconData? icon,
  }) {
    return _buildCard(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                if (icon != null) ...[
                  Icon(icon, size: 18, color: primaryColor),
                  const SizedBox(width: 8),
                ],
                Text(
                  label,
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: Colors.black54,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            TextField(
              controller: controller,
              maxLines: maxLines,
              keyboardType: keyboardType,
              decoration: InputDecoration(
                hintText: hint,
                hintStyle: TextStyle(color: Colors.grey[400], fontSize: 14),
                filled: true,
                fillColor: Colors.grey[50],
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey[300]!),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey[300]!),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: primaryColor, width: 2),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
              ),
              style: const TextStyle(fontSize: 14),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildToggleSetting({
    required String title,
    required bool value,
    required Function(bool) onChanged,
    IconData? icon,
  }) {
    return _buildCard(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            if (icon != null) ...[
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: value
                      ? primaryColor.withOpacity(0.1)
                      : Colors.grey[100],
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(
                  icon,
                  size: 20,
                  color: value ? primaryColor : Colors.grey[400],
                ),
              ),
              const SizedBox(width: 12),
            ],
            Expanded(
              child: Text(
                title,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                  color: Colors.black87,
                ),
              ),
            ),
            Switch(
              value: value,
              onChanged: onChanged,
              activeColor: primaryColor,
              activeTrackColor: primaryColor.withOpacity(0.3),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRadioGroup<T>({
    required List<RadioOption<T>> options,
    required T groupValue,
    required Function(T?) onChanged,
  }) {
    return _buildCard(
      child: Column(
        children: options.asMap().entries.map((entry) {
          final index = entry.key;
          final option = entry.value;
          final isSelected = groupValue == option.value;

          return InkWell(
            onTap: () => onChanged(option.value),
            borderRadius: BorderRadius.vertical(
              top: index == 0 ? const Radius.circular(12) : Radius.zero,
              bottom: index == options.length - 1
                  ? const Radius.circular(12)
                  : Radius.zero,
            ),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                border: index < options.length - 1
                    ? Border(bottom: BorderSide(color: Colors.grey[200]!))
                    : null,
              ),
              child: Row(
                children: [
                  Container(
                    width: 24,
                    height: 24,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: isSelected ? primaryColor : Colors.grey[400]!,
                        width: isSelected ? 6 : 2,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      option.title,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: isSelected
                            ? FontWeight.w600
                            : FontWeight.w400,
                        color: isSelected ? primaryColor : Colors.black87,
                      ),
                    ),
                  ),
                  if (isSelected)
                    Icon(Icons.check_circle, color: primaryColor, size: 20),
                ],
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: backgroundColor,
      appBar: AppBar(
        title: const Text(
          'Kelola Nota',
          style: TextStyle(
            color: Colors.black87,
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black87),
          onPressed: () => Navigator.pop(context),
        ),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(1),
          child: Container(color: Colors.grey[200], height: 1),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: primaryColor))
          : SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    'Profil Nota',
                    subtitle: 'Informasi header pada nota (dari profil outlet)',
                    icon: Icons.store,
                  ),
                  _buildReadOnlyField(
                    label: 'Nama Bisnis',
                    value: widget.outletName,
                    icon: Icons.business,
                  ),
                  _buildReadOnlyField(
                    label: 'Alamat',
                    value: widget.outletAddress,
                    icon: Icons.location_on,
                  ),
                  _buildReadOnlyField(
                    label: 'Nomor Telepon',
                    value: widget.outletPhone,
                    icon: Icons.phone,
                  ),
                  _buildSectionHeader(
                    'Footer Nota',
                    subtitle: 'Catatan di bagian bawah nota',
                    icon: Icons.note,
                  ),
                  _buildTextField(
                    controller: _footerNoteController,
                    label: 'Catatan Footer',
                    hint:
                        '• Harap membawa bon ini ketika mengambil pakaian\n• Terima kasih',
                    maxLines: 3,
                  ),
                  _buildSectionHeader(
                    'Footer Nota WhatsApp',
                    subtitle: 'Catatan untuk nota yang dikirim via WhatsApp',
                    icon: Icons.chat,
                  ),
                  _buildTextField(
                    controller: _whatsappNoteController,
                    label: 'Catatan WhatsApp',
                    hint: 'Terima kasih telah menggunakan layanan kami',
                    maxLines: 3,
                  ),
                  _buildSectionHeader(
                    'Pengaturan Tampilan',
                    subtitle: 'Atur elemen yang ditampilkan pada nota',
                    icon: Icons.visibility,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan Logo',
                    value: _showLogo,
                    onChanged: (value) => setState(() => _showLogo = value),
                    icon: Icons.image,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan QR Code',
                    value: _showQRCode,
                    onChanged: (value) => setState(() => _showQRCode = value),
                    icon: Icons.qr_code,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan Nama Toko',
                    value: _showBusinessName,
                    onChanged: (value) =>
                        setState(() => _showBusinessName = value),
                    icon: Icons.storefront,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan Keterangan',
                    value: _showDescription,
                    onChanged: (value) =>
                        setState(() => _showDescription = value),
                    icon: Icons.description,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan Footer Nota',
                    value: _showFooterNote,
                    onChanged: (value) =>
                        setState(() => _showFooterNote = value),
                    icon: Icons.note_add,
                  ),
                  _buildToggleSetting(
                    title: 'Tampilkan Footer WhatsApp',
                    value: _showWhatsappFooter,
                    onChanged: (value) =>
                        setState(() => _showWhatsappFooter = value),
                    icon: Icons.chat_bubble,
                  ),

                  _buildSectionHeader(
                    'Ukuran Printer',
                    subtitle: 'Pilih ukuran kertas printer',
                    icon: Icons.print,
                  ),
                  _buildRadioGroup<String>(
                    options: [
                      RadioOption(title: '58 mm', value: '58'),
                      RadioOption(title: '80 mm', value: '80'),
                    ],
                    groupValue: _printerSize,
                    onChanged: (value) => setState(() => _printerSize = value!),
                  ),
                  _buildSectionHeader(
                    'Tipe Printer',
                    subtitle: 'Pilih tipe format nota',
                    icon: Icons.print_outlined,
                  ),
                  _buildRadioGroup<String>(
                    options: [
                      RadioOption(
                        title: 'Type A - Format Standard',
                        value: 'A',
                      ),
                      RadioOption(
                        title: 'Type B - Format Alternatif',
                        value: 'B',
                      ),
                    ],
                    groupValue: _printerType,
                    onChanged: (value) => setState(() => _printerType = value!),
                  ),

                  const SizedBox(height: 100),
                ],
              ),
            ),
      bottomNavigationBar: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.05),
              blurRadius: 10,
              offset: const Offset(0, -2),
            ),
          ],
        ),
        child: SafeArea(
          child: ElevatedButton(
            onPressed: _isSaving ? null : _saveNotaSettings,
            style: ElevatedButton.styleFrom(
              backgroundColor: primaryColor,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              elevation: 0,
            ),
            child: _isSaving
                ? Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: const [
                      SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2,
                        ),
                      ),
                      SizedBox(width: 12),
                      Text(
                        'Menyimpan...',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  )
                : Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: const [
                      Icon(Icons.save, size: 20),
                      SizedBox(width: 8),
                      Text(
                        'SIMPAN PENGATURAN',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
          ),
        ),
      ),
    );
  }
}

class RadioOption<T> {
  final String title;
  final T value;

  RadioOption({required this.title, required this.value});
}
