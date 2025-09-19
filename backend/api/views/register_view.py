from rest_framework.views import APIView
from rest_framework import status
from rest_framework.response import Response
from django.contrib.sites.shortcuts import get_current_site
from django.utils.http import urlsafe_base64_encode, urlsafe_base64_decode
from api.tokens import account_activation_token
from django.urls import reverse
from django.template.loader import render_to_string
from django.core.mail import send_mail
from django.conf import settings
from django.utils.encoding import force_bytes, force_str
from api.models import User

class RegisterView(APIView):

	def post(self, request):
		email = request.data.get('email')
		login = request.data.get('login')
		first_name = request.data.get('first_name')
		last_name = request.data.get('last_name')
		password = request.data.get('password')
		
		if not all([email, login, first_name, last_name, password]):
			return Response({'error': 'All fields are required'}, status=status.HTTP_400_BAD_REQUEST)
		if User.objects.filter(email=email).exists() or User.objects.filter(login=login).exists():
			return Response({'error': 'Registration failed. Please try again with different credentials.'}, status=status.HTTP_400_BAD_REQUEST)
		try: 
			user = User.objects.create_user(
				login=login,
				email=email,
				first_name=first_name,
				last_name=last_name,
				password=password,
    			is_active=False,
			)
			current_site = get_current_site(request)
			uidb64 = urlsafe_base64_encode(force_bytes(user.pk))
			token = account_activation_token.make_token(user)
			activation_token = reverse('activate', kwargs={'uidb64': uidb64, 'token': token})
			scheme = 'https' if request.is_secure() else 'http'
			activation_link = f'{scheme}://{current_site.domain}{activation_token}'
			subject = 'Activate your account'
			message = render_to_string('activation_email.html', {
				'user': user,
				'link': activation_link,
			})
			plain_message = f'Hi {user.first_name},\nPlease activate your account using the following link:\n{activation_link}'
			send_mail(
			    subject,
			    plain_message,
			    settings.DEFAULT_FROM_EMAIL,
			    [user.email],
			    html_message=message,
			)

			return Response({'message': 'User registered successfully. Confirm registration via link sent to your email'}, status=status.HTTP_201_CREATED)
		except Exception as e:
			return Response({'error': 'Internal server error'}, status=status.HTTP_500_INTERNAL_SERVER_ERROR)

class ActivateAccountView(APIView):
    
	def get(self, request, uidb64, token):
		try:
			uid = force_str(urlsafe_base64_decode(uidb64))
			user = User.objects.get(pk=uid)
		except (TypeError, ValueError, OverflowError, User.DoesNotExist) as e:
			user = None
		if user is not None and account_activation_token.check_token(user, token):
			user.is_active = True
			user.save()
			return Response({'message': 'Account activated successfully'}, status=status.HTTP_200_OK)
		else:
			return Response({'error': 'Activation link is invalid'}, status=status.HTTP_400_BAD_REQUEST)