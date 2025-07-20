from django.contrib.auth.models import AbstractUser, PermissionsMixin, UserManager
from django.db import models
from django.utils import timezone

class CustomUserManager(UserManager):
    def get_user_by_login(self, login):
        try:
            return self.get(student_login=login)
        except self.model.DoesNotExist:
            return None
    def get_users_logins(self):
        """
        Retrieves all users' logins.
        Returns:
            list: A list of all users' logins.
        """
        return self.values_list('student_login', flat=True)

class User(AbstractUser, PermissionsMixin):
    id = models.BigAutoField(primary_key=True)
    login = models.CharField(max_length=100, unique=True)
    avatar = models.CharField(max_length=100, blank=True, help_text="Filename of the avatar stored in media/avatars/")

    # Add fields required for Django auth
    is_active = models.BooleanField(default=True)
    is_staff = models.BooleanField(default=False)
    date_joined = models.DateTimeField(default=timezone.now)
    
    objects = CustomUserManager()

    USERNAME_FIELD = 'login'
    REQUIRED_FIELDS = ['id']

    def __str__(self):
        return self.login
