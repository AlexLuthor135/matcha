class WebhookAuthError(Exception):
    """Webhook could not be validated with secret."""
    def __init__(self, message="Couldn't validate webhook secret", status=403):
        super().__init__(message)
        self.message = message
        self.status = status

class WebhookFormatError(Exception):
    """Bad format in WebhookRequest."""
    def __init__(self, message="Wrong format in webhook", status=400):
        super().__init__(message)
        self.message = message
        self.status = status
