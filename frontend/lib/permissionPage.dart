import 'package:flutter/material.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';
import 'dart:io';

class PermissionPage extends StatefulWidget {
  const PermissionPage({super.key});

  @override
  State<PermissionPage> createState() => _PermissionPageState();
}

class _PermissionPageState extends State<PermissionPage> {
  Map<String, String> permissions = {
    'Bluetooth Connection': 'Tidak Diizinkan',
    'Bluetooth Scan': 'Tidak Diizinkan',
    'Fine Location': 'Tidak Diizinkan',
    'Os Android Version': '13',
    'Actual Level Version': '33',
  };

  @override
  void initState() {
    super.initState();
    _checkPermissions();
    _getAndroidVersion();
  }

  Future<void> _getAndroidVersion() async {
    if (Platform.isAndroid) {
      setState(() {
        permissions['Os Android Version'] = '13';
        permissions['Actual Level Version'] = '33';
      });
    }
  }

  Future<void> _checkPermissions() async {
    final btConnect = await Permission.bluetoothConnect.status;
    final btScan = await Permission.bluetoothScan.status;
    final location = await Permission.locationWhenInUse.status;

    setState(() {
      permissions['Bluetooth Connection'] = 
          btConnect.isGranted ? 'Diizinkan' : 'Tidak Diizinkan';
      permissions['Bluetooth Scan'] = 
          btScan.isGranted ? 'Diizinkan' : 'Tidak Diizinkan';
      permissions['Fine Location'] = 
          location.isGranted ? 'Diizinkan' : 'Tidak Diizinkan';
    });
  }

  Future<void> _requestBluetoothPermission() async {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Meminta izin Bluetooth...'),
        backgroundColor: Colors.blue,
        duration: Duration(seconds: 1),
      ),
    );

    Map<Permission, PermissionStatus> statuses = await [
      Permission.bluetoothConnect,
      Permission.bluetoothScan,
      Permission.locationWhenInUse,
    ].request();

    await _checkPermissions();

    if (mounted) {
      bool allGranted = statuses.values.every((status) => status.isGranted);
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            allGranted 
                ? 'Semua permission berhasil diberikan!' 
                : 'Beberapa permission ditolak'
          ),
          backgroundColor: allGranted ? Colors.green : Colors.orange,
        ),
      );
    }
  }

  Future<void> _scanBluetooth() async {
    // Cek permission dulu
    final btScan = await Permission.bluetoothScan.status;
    final location = await Permission.locationWhenInUse.status;

    if (!btScan.isGranted || !location.isGranted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Permission belum diberikan!'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Memulai scan Bluetooth...'),
        backgroundColor: Colors.blue,
        duration: Duration(seconds: 2),
      ),
    );

    try {
      // Mulai scan
      await FlutterBluePlus.startScan(timeout: const Duration(seconds: 4));
      
      // Listen hasil scan
      FlutterBluePlus.scanResults.listen((results) {
        for (ScanResult r in results) {
          print('Device found: ${r.device.name} - ${r.device.id}');
        }
      });

      // Stop scan setelah timeout
      await Future.delayed(const Duration(seconds: 4));
      await FlutterBluePlus.stopScan();

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Scan selesai!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'Kembali',
          style: TextStyle(
            color: Colors.black,
            fontSize: 16,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.05),
                    blurRadius: 10,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Column(
                children: [
                  _buildPermissionItem(
                    'Bluetooth Connection',
                    permissions['Bluetooth Connection']!,
                    isRed: permissions['Bluetooth Connection'] == 'Tidak Diizinkan',
                  ),
                  _buildDivider(),
                  _buildPermissionItem(
                    'Bluetooth Scan',
                    permissions['Bluetooth Scan']!,
                    isRed: permissions['Bluetooth Scan'] == 'Tidak Diizinkan',
                  ),
                  _buildDivider(),
                  _buildPermissionItem(
                    'Fine Location',
                    permissions['Fine Location']!,
                    isRed: permissions['Fine Location'] == 'Tidak Diizinkan',
                  ),
                  _buildDivider(),
                  _buildPermissionItem(
                    'Os Android Version',
                    permissions['Os Android Version']!,
                  ),
                  _buildDivider(),
                  _buildPermissionItem(
                    'Actual Level Version',
                    permissions['Actual Level Version']!,
                    isLast: true,
                  ),
                ],
              ),
            ),
            
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _requestBluetoothPermission,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFE85C40),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 0,
                ),
                child: const Text(
                  'BT Connection Permission',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),

            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _scanBluetooth,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFE85C40),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 0,
                ),
                child: const Text(
                  'BT Scan',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPermissionItem(String title, String value, {bool isRed = false, bool isLast = false}) {
    return Padding(
      padding: EdgeInsets.fromLTRB(20, 16, 20, isLast ? 16 : 0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            title,
            style: TextStyle(
              fontSize: 13,
              color: isRed ? Colors.red[300] : Colors.black87,
              fontWeight: FontWeight.w400,
            ),
          ),
          Text(
            value,
            style: TextStyle(
              fontSize: 13,
              color: isRed ? Colors.red[300] : Colors.black87,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDivider() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
      child: Divider(height: 1, color: Colors.grey[200]),
    );
  }
}