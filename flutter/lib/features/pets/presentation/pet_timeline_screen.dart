import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

class PetTimelineScreen extends StatefulWidget {
  const PetTimelineScreen({super.key, required this.petId});
  final String petId;

  @override
  State<PetTimelineScreen> createState() => _PetTimelineScreenState();
}

class _PetTimelineScreenState extends State<PetTimelineScreen> {
  List<dynamic> items = [];

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    final data = await ApiClient.instance.getTimeline(widget.petId);
    setState(() => items = data);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Historique')),
      body: ListView.builder(
        itemCount: items.length,
        itemBuilder: (_, i) {
          final item = items[i] as Map<String, dynamic>;
          return ListTile(
            title: Text(item['title'] as String? ?? ''),
            subtitle: Text(item['body'] as String? ?? ''),
          );
        },
      ),
    );
  }
}
