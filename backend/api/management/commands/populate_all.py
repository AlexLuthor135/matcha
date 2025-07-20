from django.core.management.base import BaseCommand
from django.core.management import call_command

class Command(BaseCommand):
    help = "Run all setup/population commands in sequence."

    def handle(self, *args, **options):
        self.stdout.write(self.style.NOTICE("Starting unified activation process..."))
        try:
            self.stdout.write(self.style.NOTICE("1. Populating users..."))
            call_command('populate_users')

            self.stdout.write(self.style.NOTICE("2. Populating user locations..."))
            call_command('populate_locations')
            
            self.stdout.write(self.style.NOTICE("3. Populating ranks..."))
            call_command('populate_ranks')

            self.stdout.write(self.style.NOTICE("4. Populating friends..."))
            call_command('populate_friends')

            self.stdout.write(self.style.NOTICE("5. Populating happeer scores..."))
            call_command('populate_happeer')
            
            self.stdout.write(self.style.NOTICE("6. Populating badpeer scores..."))
            call_command('populate_badpeer')
            
            self.stdout.write(self.style.NOTICE("6. Populating avatars..."))
            call_command('populate_avatars')

            self.stdout.write(self.style.SUCCESS("✅ All commands executed successfully."))
        except Exception as e:
            self.stderr.write(self.style.ERROR(f"❌ Error occurred: {e}"))
