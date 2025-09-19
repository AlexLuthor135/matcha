from django.contrib.auth.models import AbstractUser, UserManager
from django.db import models
from django.utils import timezone

class CustomUserManager(UserManager):
    def create_user(self, login, password=None, **extra_fields):
        pass
        if not login:
            raise ValueError('The Login field must be set')
        user = self.model(student_login=login, **extra_fields)
        user.set_password(password)
        user.save(using=self._db)
        return user
    
    def create_superuser(self, login, password=None, **extra_fields):
        extra_fields.setdefault('is_staff', True)
        extra_fields.setdefault('is_superuser', True)

        if extra_fields.get('is_staff') is not True:
            raise ValueError('Superuser must have is_staff=True.')
        if extra_fields.get('is_superuser') is not True:
            raise ValueError('Superuser must have is_superuser=True.')
        return self.create_user(login, password, **extra_fields)
    
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

class User(AbstractUser):
    id = models.BigAutoField(primary_key=True)
    login = models.CharField(max_length=100, unique=True)
    password = models.CharField(max_length=128)
    first_name = models.CharField(max_length=30)
    last_name = models.CharField(max_length=30)
    email = models.EmailField(unique=True)
    avatar = models.CharField(max_length=100, blank=True, help_text="Filename of the avatar stored in media/avatars/")

    is_active = models.BooleanField(default=True)
    is_staff = models.BooleanField(default=False)
    date_joined = models.DateTimeField(default=timezone.now)
    
    objects = CustomUserManager()

    USERNAME_FIELD = 'login'
    REQUIRED_FIELDS = ['id']
        
    def __str__(self):
        return self.login
