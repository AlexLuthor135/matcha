import os
import requests
from rest_framework.response import Response
from intra_client.services.intra_api import ic
from intra_client.services.intra_utils import *
from django.conf import settings
from api.models import User
from concurrent.futures import ThreadPoolExecutor
from django.db import close_old_connections
from intra_client.services.logging import peersphere_logger

AVATAR_DIR = os.path.join(settings.MEDIA_ROOT, 'avatars')
os.makedirs(AVATAR_DIR, exist_ok=True)

def get_avatar(student_login):
    try:
        user_data = get_user_info(ic, student_login)
        image_url = user_data.get('image', {}).get('versions', {}).get('medium')
        if not image_url:
            peersphere_logger.log_admins(f"❌ No image URL found for {student_login}")
            return None

        filename = f"{student_login}.jpg"
        file_path = os.path.join(AVATAR_DIR, filename)
        response = requests.get(image_url, stream=True)
        if response.status_code != 200:
            peersphere_logger.log_admins(f"❌ Failed to download avatar for {student_login}")
            return None
        new_image_data = b''.join(response.iter_content(8192))
        new_image_size = len(new_image_data)

        if os.path.exists(file_path):
            existing_image_size = os.path.getsize(file_path)
            if new_image_size == existing_image_size:
                peersphere_logger.log_admins(f"⚠️ Avatar already up-to-date for {student_login}. Skipping.")
                return filename
        with open(file_path, 'wb') as f:
            f.write(new_image_data)
        peersphere_logger.log_admins(f"✅ Saved/Updated avatar for {student_login}")
        return filename

    except Exception as e:
        print(f"Error processing {student_login}: {e}")
        peersphere_logger.log_admins(f"❌ Error processing {student_login}: {e}")
        return None
    
def process_user(user):
    try:
        close_old_connections()
        avatar_filename = f"{user.student_login}.jpg"
        avatar_filename = get_avatar(user.student_login)
        if avatar_filename:
            user.avatar = avatar_filename
            user.save()
        else:
            peersphere_logger.log_admins(f"❌ No avatar found for {user.student_login}")
    except Exception as e:
        peersphere_logger.log_admins(f"❌ Error processing {user.student_login}: {e}")

def populate_avatars():
    users = list(User.objects.all())
    with ThreadPoolExecutor(max_workers=4) as executor:
        executor.map(process_user, users)
    return Response({"message": "Avatars populated successfully."})