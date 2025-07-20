from django.urls import path, include
from api.views.api_view import User42LoginView, User42CallbackView, VerifyLoginView, CookieTokenRefreshView, LogoutView
from api.views.user_view import ResetBlockedView, SetBlockedView

urlpatterns = [
    # OAuth routes
    path('accounts/42/login/', User42LoginView.as_view(), name='42_login'),
    path('accounts/42/callback/', User42CallbackView.as_view(), name='42_callback'),
    path('accounts/42/logout/', LogoutView.as_view(), name='42_logout'),
    
    # Token refresh
    path('accounts/verify_login/', VerifyLoginView.as_view(), name='token_refresh'),
    path('accounts/token/refresh/', CookieTokenRefreshView.as_view(), name='cookie_token_refresh'),
    path('accounts/', include('allauth.urls')),
    
    # Eval set/reset
    path('user/reset_blocked/', ResetBlockedView.as_view(), name='reset_blocked'),
    path('user/set_blocked/', SetBlockedView.as_view(), name='set_blocked'),
]