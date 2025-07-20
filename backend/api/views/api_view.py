from django.conf import settings
from requests_oauthlib import OAuth2Session
import requests
import secrets
from django.http import JsonResponse, HttpResponseRedirect
from django.shortcuts import redirect
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework.permissions import IsAuthenticated
from rest_framework_simplejwt.tokens import RefreshToken
from django.contrib.auth import login
from django.views.decorators.csrf import csrf_exempt
from django.utils.decorators import method_decorator
from django.core.cache import cache
from api.models import User
from requests.exceptions import RequestException
from urllib.parse import urlencode

@method_decorator(csrf_exempt, name='dispatch')
class User42LoginView(APIView):
    def get(self, request):
        try:
            oauth = OAuth2Session(
                settings.CLIENT_ID,
                redirect_uri=settings.REDIRECT_URI,
                scope=settings.OAUTH_SCOPE
            )
            
            state = secrets.token_urlsafe(16)
            cache.set(f'oauth_state_{state}', state, timeout=300)
            
            authorization_url, state = oauth.authorization_url(
                settings.AUTHORIZATION_BASE_URL,
                state=state
            )
            
            return redirect(authorization_url)
            
        except Exception as e:
            return JsonResponse({
                'error': 'Authorization initialization failed',
                'detail': str(e)
            }, status=500)

class User42CallbackView(APIView):
    def get(self, request):
        try:
            state = request.GET.get('state')
            stored_state = cache.get(f'oauth_state_{state}')
            
            if not state or state != stored_state:
                return JsonResponse({
                    'error': 'Invalid state parameter'
                }, status=400)
                
            cache.delete(f'oauth_state_{state}')

            oauth = OAuth2Session(
                settings.CLIENT_ID, 
                state=request.session.get('oauth_state'),
                redirect_uri=settings.REDIRECT_URI
            )
            
            token = oauth.fetch_token(
                settings.TOKEN_URL,
                client_secret=settings.CLIENT_SECRET,
                authorization_response=request.build_absolute_uri(),
                include_client_id=True
            )

            if not token or 'access_token' not in token:
                raise ValueError('Invalid token response')
            
            response = requests.get(
                settings.USER_INFO_URL,
                headers={'Authorization': f'Bearer {token["access_token"]}'},
                timeout=10
            )
            response.raise_for_status()
            user_info = response.json()
            
            required_fields = ['id']
            if not all(field in user_info for field in required_fields):
                raise ValueError('Missing required user information')
            
            student_login = user_info['login']
            if not student_login:
                return JsonResponse({'error': 'Invalid user data: login is missing'}, status=400)
            try:
                user = User.objects.get(student_login=student_login)
            except User.DoesNotExist:
                error_url = f"{settings.FRONTEND_URL}/login/callback?{urlencode({'error': 'User not registered in the system'})}"
                return redirect(error_url)
            
            login(request, user)

            refresh = RefreshToken.for_user(user)
            access_token = str(refresh.access_token)
            refresh_token = str(refresh)

            response = HttpResponseRedirect(f"{settings.FRONTEND_URL}/login/callback")
            response.set_cookie("access_token", access_token, httponly=True, secure=True, samesite="Lax")
            response.set_cookie("refresh_token", refresh_token, httponly=True, secure=True, samesite="Lax")

            return response
        except RequestException as e:
            print(f"Request error occurred: {e}")
            return redirect(settings.FRONTEND_URL)
        except Exception as e:
            print(f"An error occurred: {e}")
            return redirect(settings.FRONTEND_URL)

class VerifyLoginView(APIView):
    permission_classes = [IsAuthenticated]

    def get(self, request):
        return Response({
            "id": request.user.id,
            "student_login": request.user.student_login,
            "student_id": request.user.student_id,
            "is_staff": request.user.is_staff,
        })

@method_decorator(csrf_exempt, name='dispatch')
class CookieTokenRefreshView(APIView):
    def post(self, request):
        refresh_token = request.COOKIES.get('refresh_token')
        if not refresh_token:
            return JsonResponse({'error': 'No refresh token provided'}, status=401)
        try:
            refresh = RefreshToken(refresh_token)
            access = refresh.access_token
            response = JsonResponse({'message': 'Token refreshed'})
            response.set_cookie(
                "access_token",
                str(access),
                httponly=True,
                secure=True,
                samesite="Lax"
            )
            return response
        except Exception as e:
            return JsonResponse({'error': str(e)}, status=401)
        
class LogoutView(APIView):
    permission_classes = [IsAuthenticated]
    
    def post(self, request):
        try:
            response = JsonResponse({'message': 'Logged out successfully'})
            response.delete_cookie('access_token')
            response.delete_cookie('refresh_token')
            return response
        except Exception as e:
            return JsonResponse({'error': str(e)}, status=500)