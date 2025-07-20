from rest_framework import serializers
from api.models import User

class UserSerializer(serializers.ModelSerializer):
    class Meta:
        model = User
        fields = ['id', 'student_login', 'avatar', 'quest_rank', 'happeer_score', 'badpeer_score', 'has_eval', 'blocked_until', 'weekly_evals',]
