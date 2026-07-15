import 'package:flutter/material.dart';
import 'package:petsfollow_mobile/core/api/api_client.dart';

class MessagingScreen extends StatefulWidget {
  const MessagingScreen({super.key});

  @override
  State<MessagingScreen> createState() => _MessagingScreenState();
}

class _MessagingScreenState extends State<MessagingScreen> {
  List<dynamic> messages = [];
  String? threadId;
  final draft = TextEditingController();

  @override
  void initState() {
    super.initState();
    initThread();
  }

  Future<void> initThread() async {
    final threads = await ApiClient.instance.getThreads();
    if (threads.isEmpty) return;
    threadId = (threads.first as Map<String, dynamic>)['id'] as String?;
    await loadMessages();
  }

  Future<void> loadMessages() async {
    if (threadId == null) return;
    final data = await ApiClient.instance.getMessages(threadId!);
    setState(() => messages = data);
  }

  Future<void> send() async {
    if (threadId == null || draft.text.isEmpty) return;
    await ApiClient.instance.sendMessage(threadId!, draft.text);
    draft.clear();
    await loadMessages();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Messagerie véto')),
      body: Column(children: [
        Expanded(
          child: ListView.builder(
            itemCount: messages.length,
            itemBuilder: (_, i) {
              final m = messages[i] as Map<String, dynamic>;
              return ListTile(title: Text(m['body'] as String? ?? ''));
            },
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(8),
          child: Row(children: [
            Expanded(child: TextField(controller: draft)),
            IconButton(onPressed: send, icon: const Icon(Icons.send)),
          ]),
        ),
      ]),
    );
  }
}
