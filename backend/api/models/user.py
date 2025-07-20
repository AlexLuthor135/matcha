from django.contrib.auth.models import AbstractUser, PermissionsMixin, UserManager
from django.db import models
from django.utils import timezone
from datetime import timedelta
from .evaluation import Evaluation

class CustomUserManager(UserManager):
    def get_user_by_login(self, login):
        try:
            return self.get(student_login=login)
        except self.model.DoesNotExist:
            return None
    
    def get_users_by_rank(self, rank_number):
        """
        Retrieves all users with the specified rank.
        Args:
            rank_number (int): The rank to filter users by.
        Returns:
            QuerySet: A queryset of users with the specified rank.
        """
        return self.filter(quest_rank=rank_number)
    
    def get_users_rank(self, login):
        """
        Retrieves a user's rank.
        Args:
            login (str): The login of the user to retrieve the rank for.
        Returns:
            int: The user's rank, or None if the user does not exist.
        """
        user = self.get_user_by_login(login)
        return getattr(user, 'quest_rank', None)

    def get_users_logins(self):
        """
        Retrieves all users' logins.
        Returns:
            list: A list of all users' logins.
        """
        return self.values_list('student_login', flat=True)

    def get_active_students(self):
        pass
        """
        Retrieves logins of all active students.
        Returns:
            QuerySet: A queryset of active students.
        """
        return self.filter(has_location=True).values_list('student_login', flat=True)
    
    def get_friends(self, login):
        """
        Retrieves a user's friends.
        Args:
            login (str): The login of the user to retrieve friends for.
        Returns:
            list: A list of the user's friends.
        """
        user = self.get_user_by_login(login)
        return getattr(user, 'friends', [])
        
    def get_user_eval_status(self, login):
        """
        Returns True if the user is currently under evaluation or temporarily blocked from evaluations.
        """
        user = self.get_user_by_login(login)
        if not user or user.available_pass:
            return False
        is_currently_evaluated = user.has_eval
        is_temporarily_blocked = user.blocked_until and user.blocked_until > timezone.now()
        has_reached_weekly_limit = (user.weekly_evals or 0) >= 3

        return is_currently_evaluated or is_temporarily_blocked or has_reached_weekly_limit

class User(AbstractUser, PermissionsMixin):
    id = models.BigAutoField(primary_key=True)
    student_id = models.IntegerField(unique=True)
    student_login = models.CharField(max_length=100, unique=True)
    happeer_score = models.IntegerField(default=0)
    badpeer_score = models.IntegerField(default=0)
    cancelled_evals = models.IntegerField(default=0)
    quest_rank = models.IntegerField(null=True, blank=True)
    last_login = models.DateTimeField(blank=True, null=True)
    last_campus_login = models.DateTimeField(blank=True, null=True)
    last_campus_logout = models.DateTimeField(blank=True, null=True)
    has_location = models.BooleanField(default=False)
    has_eval = models.BooleanField(default=False)
    blocked_until = models.DateTimeField(null=True, blank=True)
    block_pass_used_at = models.DateTimeField(null=True, blank=True)
    available_pass = models.BooleanField(default=False)
    friends = models.JSONField(default=list, blank=True)
    avatar = models.CharField(max_length=100, blank=True, help_text="Filename of the avatar stored in media/avatars/")

    # Add fields required for Django auth
    is_active = models.BooleanField(default=True)
    is_staff = models.BooleanField(default=False)
    date_joined = models.DateTimeField(default=timezone.now)
    
    objects = CustomUserManager()

    USERNAME_FIELD = 'student_login'
    REQUIRED_FIELDS = ['student_id']

    def __str__(self):
        return self.student_login
    
    @property
    def weekly_evals(self):
        one_week_ago = timezone.now() - timedelta(days=7)
        return Evaluation.objects.filter(evaluator=self.student_login, time_created__gte=one_week_ago,).count()
    
    @property
    def block_pass(self):
        if not self.block_pass_used_at:
            return True
        return self.block_pass_used_at + timedelta(days=7) <= timezone.now()
