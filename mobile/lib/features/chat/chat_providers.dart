import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/api_client.dart';
import 'data/chat_repository_impl.dart';
import 'domain/chat_repository.dart';
import 'domain/models.dart';

/// Provides the singleton [ApiClient]. Declared here to avoid a circular
/// import; re-exported from core/di.dart.
final apiClientProvider = Provider<ApiClient>((ref) {
  final client = ApiClient();
  ref.onDispose(client.dispose);
  return client;
});

/// Provides the [ChatRepository] implementation wired to [apiClientProvider].
final chatRepositoryProvider = Provider<ChatRepository>((ref) {
  return ChatRepositoryImpl(apiClient: ref.watch(apiClientProvider));
});

/// State holder for chat screen.
class ChatState {
  const ChatState({
    this.isLoading = false,
    this.result,
    this.errorMessage,
  });

  final bool isLoading;
  final ChatResult? result;
  final String? errorMessage;

  ChatState copyWith({
    bool? isLoading,
    ChatResult? result,
    String? errorMessage,
    bool clearResult = false,
    bool clearError = false,
  }) {
    return ChatState(
      isLoading: isLoading ?? this.isLoading,
      result: clearResult ? null : (result ?? this.result),
      errorMessage: clearError ? null : (errorMessage ?? this.errorMessage),
    );
  }
}

/// Notifier that drives the chat screen state machine.
class ChatNotifier extends StateNotifier<ChatState> {
  ChatNotifier(this._repository) : super(const ChatState());

  final ChatRepository _repository;

  Future<void> sendQuery(String question, LedgerSummary ledger) async {
    if (question.trim().length < 3) return;

    state = state.copyWith(
      isLoading: true,
      clearResult: true,
      clearError: true,
    );

    try {
      final result = await _repository.query(question.trim(), ledger);
      state = state.copyWith(isLoading: false, result: result);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString(),
        clearResult: true,
      );
    }
  }

  void reset() => state = const ChatState();
}

final chatNotifierProvider =
    StateNotifierProvider<ChatNotifier, ChatState>((ref) {
  return ChatNotifier(ref.watch(chatRepositoryProvider));
});
