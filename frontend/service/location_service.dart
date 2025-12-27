import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/region.dart';

class LocationService {
  static const baseUrl = 'https://www.emsifa.com/api-wilayah-indonesia/api';
  static Future<List<Region>> getProvinces() async {
    final res = await http.get(Uri.parse('$baseUrl/provinces.json'));
    final data = jsonDecode(res.body) as List;
    return data.map((e) => Region.fromJson(e)).toList();
  }

  static Future<List<Region>> getCities(String provId) async {
    final res = await http.get(Uri.parse('$baseUrl/regencies/$provId.json'));
    final data = jsonDecode(res.body) as List;
    return data.map((e) => Region.fromJson(e)).toList();
  }

  static Future<List<Region>> getDistricts(String cityId) async {
    final res = await http.get(Uri.parse('$baseUrl/districts/$cityId.json'));
    final data = jsonDecode(res.body) as List;
    return data.map((e) => Region.fromJson(e)).toList();
  }
}
