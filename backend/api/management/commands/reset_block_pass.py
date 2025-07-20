from django.core.management.base import BaseCommand
from api.models import User
from intra_client.services.logging import peersphere_logger

class Command(BaseCommand):
    help = 'Reset block_pass for all users'

    def handle(self, *args, **kwargs):
        try:
            updated = User.objects.update(block_pass=True)
            peersphere_logger.log_reset(f"✅ Reset block_pass for {updated} users")
        except Exception as e:
            peersphere_logger.log_reset(f"❌ Error resetting block_pass: {e}")
