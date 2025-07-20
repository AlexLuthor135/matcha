from django.urls import path, include
from api.views.api_view import User42LoginView, User42CallbackView, VerifyLoginView, CookieTokenRefreshView, LogoutView

urlpatterns = [
    # OAuth routes
    path('accounts/42/login/', User42LoginView.as_view(), name='42_login'),
    path('accounts/42/callback/', User42CallbackView.as_view(), name='42_callback'),
    path('accounts/42/logout/', LogoutView.as_view(), name='42_logout'),
    
    # Token refresh
    path('accounts/verify_login/', VerifyLoginView.as_view(), name='token_refresh'),
    path('accounts/token/refresh/', CookieTokenRefreshView.as_view(), name='cookie_token_refresh'),
    path('accounts/', include('allauth.urls')),
]