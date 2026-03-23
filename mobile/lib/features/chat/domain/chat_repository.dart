import 'models.dart';

/// Abstract interface for chat queries against the edge-api.
abstract class ChatRepository {
  /// Sends [question] with [ledger] context to the backend and returns a
  /// [ChatResult] containing the answer and metric citations.
  Future<ChatResult> query(String question, LedgerSummary ledger);
}
