import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ai_budget/features/chat/chat_providers.dart';
import 'package:ai_budget/features/chat/domain/chat_repository.dart';
import 'package:ai_budget/features/chat/domain/models.dart';
import 'package:ai_budget/features/chat/presentation/chat_screen.dart';

/// Minimal in-memory mock that returns a canned answer.
class _MockChatRepository implements ChatRepository {
  @override
  Future<ChatResult> query(String question, LedgerSummary ledger) async {
    return ChatResult(
      answer: 'You spent \$42 on groceries.',
      citations: [
        const MetricCitation(label: 'Category', value: 'Groceries'),
        const MetricCitation(label: 'Total', value: '\$42.00'),
      ],
    );
  }
}

void main() {
  testWidgets('ChatScreen renders input field and send button',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          chatRepositoryProvider
              .overrideWithValue(_MockChatRepository()),
        ],
        child: const MaterialApp(
          home: ChatScreen(),
        ),
      ),
    );

    // Input field is present
    expect(find.byType(TextField), findsOneWidget);

    // Send button is present
    expect(find.byTooltip('Send'), findsOneWidget);

    // Placeholder hint text
    expect(find.text('Ask about your spending…'), findsOneWidget);
  });

  testWidgets('ChatScreen shows answer and citation chips after query',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          chatRepositoryProvider
              .overrideWithValue(_MockChatRepository()),
        ],
        child: const MaterialApp(
          home: ChatScreen(),
        ),
      ),
    );

    // Type a question
    await tester.enterText(find.byType(TextField), 'How much on groceries?');
    await tester.tap(find.byTooltip('Send'));

    // Loading indicator appears
    await tester.pump();
    expect(find.byType(CircularProgressIndicator), findsWidgets);

    // Wait for async response
    await tester.pumpAndSettle();

    // Answer is shown
    expect(find.text('You spent \$42 on groceries.'), findsOneWidget);

    // Citations are shown as chips
    expect(find.text('Category: Groceries'), findsOneWidget);
    expect(find.text('Total: \$42.00'), findsOneWidget);
  });

  testWidgets('ChatScreen does not submit question shorter than 3 chars',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          chatRepositoryProvider
              .overrideWithValue(_MockChatRepository()),
        ],
        child: const MaterialApp(
          home: ChatScreen(),
        ),
      ),
    );

    await tester.enterText(find.byType(TextField), 'hi');
    await tester.tap(find.byTooltip('Send'));
    await tester.pumpAndSettle();

    // No answer should appear
    expect(find.text('You spent \$42 on groceries.'), findsNothing);
  });
}
