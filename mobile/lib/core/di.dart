/// Dependency injection wiring for all features.
///
/// The [apiClientProvider] is defined in chat_providers.dart (to avoid
/// circular imports) and re-exported here for convenience.
library di;

export '../features/chat/chat_providers.dart'
    show apiClientProvider, chatRepositoryProvider, chatNotifierProvider;
