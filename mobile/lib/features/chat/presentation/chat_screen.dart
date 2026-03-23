import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../chat_providers.dart';
import '../domain/models.dart';

/// A read-only analytics Q&A screen.
///
/// The user types a question; the answer is shown with cited metric chips.
class ChatScreen extends ConsumerStatefulWidget {
  const ChatScreen({super.key});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _controller = TextEditingController();
  final _focusNode = FocusNode();

  // Placeholder ledger — in a real implementation this would come from the
  // transactions repository (e.g. via a Riverpod provider).
  static final _placeholderLedger = LedgerSummary(
    rangeStart: '2024-01-01',
    rangeEnd: '2024-01-31',
    totalsByCategory: const {},
  );

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _submit() {
    final question = _controller.text.trim();
    if (question.length < 3) return;
    ref
        .read(chatNotifierProvider.notifier)
        .sendQuery(question, _placeholderLedger);
    _controller.clear();
    _focusNode.unfocus();
  }

  @override
  Widget build(BuildContext context) {
    final chatState = ref.watch(chatNotifierProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Chat'),
      ),
      body: Column(
        children: [
          Expanded(
            child: _ChatBody(state: chatState),
          ),
          _InputBar(
            controller: _controller,
            focusNode: _focusNode,
            isLoading: chatState.isLoading,
            onSubmit: _submit,
          ),
        ],
      ),
    );
  }
}

class _ChatBody extends StatelessWidget {
  const _ChatBody({required this.state});

  final ChatState state;

  @override
  Widget build(BuildContext context) {
    if (state.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Text(
            state.errorMessage!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
            textAlign: TextAlign.center,
          ),
        ),
      );
    }

    if (state.result == null) {
      return const Center(
        child: Text(
          'Ask a question about your spending.',
          style: TextStyle(color: Colors.grey),
        ),
      );
    }

    return _AnswerView(result: state.result!);
  }
}

class _AnswerView extends StatelessWidget {
  const _AnswerView({required this.result});

  final ChatResult result;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            result.answer,
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          if (result.citations.isNotEmpty) ...[
            const SizedBox(height: 16),
            Text(
              'Sources',
              style: Theme.of(context).textTheme.labelLarge,
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: result.citations
                  .map((c) => _CitationChip(citation: c))
                  .toList(),
            ),
          ],
        ],
      ),
    );
  }
}

class _CitationChip extends StatelessWidget {
  const _CitationChip({required this.citation});

  final MetricCitation citation;

  @override
  Widget build(BuildContext context) {
    return Chip(
      label: Text('${citation.label}: ${citation.value}'),
      backgroundColor:
          Theme.of(context).colorScheme.secondaryContainer,
      labelStyle: TextStyle(
        color: Theme.of(context).colorScheme.onSecondaryContainer,
        fontSize: 12,
      ),
    );
  }
}

class _InputBar extends StatelessWidget {
  const _InputBar({
    required this.controller,
    required this.focusNode,
    required this.isLoading,
    required this.onSubmit,
  });

  final TextEditingController controller;
  final FocusNode focusNode;
  final bool isLoading;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 8, 8),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                focusNode: focusNode,
                decoration: const InputDecoration(
                  hintText: 'Ask about your spending…',
                  border: OutlineInputBorder(),
                  isDense: true,
                  contentPadding:
                      EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                textInputAction: TextInputAction.send,
                onSubmitted: (_) => onSubmit(),
                enabled: !isLoading,
              ),
            ),
            const SizedBox(width: 8),
            IconButton.filled(
              onPressed: isLoading ? null : onSubmit,
              icon: isLoading
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.send),
              tooltip: 'Send',
            ),
          ],
        ),
      ),
    );
  }
}
