import os
import requests
from django.core.management.base import BaseCommand
from django.db import transaction
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.conf import settings
from api.models import User

class Command(BaseCommand):
    help = "Download and store avatars from the 42 API"

    def handle(self, *args, **kwargs):
        avatar_dir = os.path.join(settings.MEDIA_ROOT, 'avatars')
        self.stdout.write(f"Saving avatars to: {avatar_dir}")
        os.makedirs(avatar_dir, exist_ok=True)
        with transaction.atomic():
            for user in User.objects.all():
                try:
                    user_data = get_user_info(ic, user.student_login)
                    image_url = user_data['image']['versions']['medium']
                    filename = f"{user.student_login}.jpg"
                    file_path = os.path.join(avatar_dir, filename)
                    if not os.path.exists(file_path):
                        response = requests.get(image_url, stream=True)
                        if response.status_code == 200:
                            with open(file_path, 'wb') as f:
                                for chunk in response.iter_content(8192):
                                    f.write(chunk)
                            self.stdout.write(self.style.SUCCESS(f"Saved avatar for {user.student_login}"))
                        else:
                            self.stdout.write(self.style.WARNING(f"Failed to download for {user.student_login}"))
                    else:
                        self.stdout.write(f"Avatar already exists for {user.student_login}")
                    user.avatar = filename
                    user.save()
                except Exception as e:
                    self.stdout.write(self.style.ERROR(f"Error processing {user.student_login}: {e}"))
