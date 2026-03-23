/// Summarised ledger context sent to the backend for chat queries.
///
/// Field names match the OpenAPI schema exactly (camelCase).
class LedgerSummary {
  const LedgerSummary({
    required this.rangeStart,
    required this.rangeEnd,
    required this.totalsByCategory,
  });

  /// ISO-8601 date string, e.g. "2024-01-01".
  final String rangeStart;

  /// ISO-8601 date string, e.g. "2024-01-31".
  final String rangeEnd;

  /// Mapping of category name to total spend amount.
  final Map<String, double> totalsByCategory;

  Map<String, dynamic> toJson() => {
        'rangeStart': rangeStart,
        'rangeEnd': rangeEnd,
        'totalsByCategory': totalsByCategory,
      };
}

/// A single cited metric returned alongside a chat answer.
class MetricCitation {
  const MetricCitation({required this.label, required this.value});

  final String label;
  final String value;

  factory MetricCitation.fromJson(Map<String, dynamic> json) =>
      MetricCitation(
        label: json['label'] as String,
        value: json['value'] as String,
      );
}

/// The result of a chat query.
class ChatResult {
  const ChatResult({required this.answer, required this.citations});

  final String answer;
  final List<MetricCitation> citations;

  factory ChatResult.fromJson(Map<String, dynamic> json) => ChatResult(
        answer: json['answer'] as String,
        citations: (json['citations'] as List<dynamic>)
            .map((e) => MetricCitation.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}
