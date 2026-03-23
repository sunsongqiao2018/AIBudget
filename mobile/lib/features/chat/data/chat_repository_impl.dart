import '../domain/chat_repository.dart';
import '../domain/models.dart';
import '../../../core/api_client.dart';

class ChatRepositoryImpl implements ChatRepository {
  const ChatRepositoryImpl({required ApiClient apiClient})
      : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<ChatResult> query(String question, LedgerSummary ledger) async {
    final responseJson = await _apiClient.post(
      '/v1/ai/chat/query',
      {
        'question': question,
        'ledgerSummary': ledger.toJson(),
      },
    );
    return ChatResult.fromJson(responseJson);
  }
}
