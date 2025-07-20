from django.contrib import admin
from django.urls import path, include


urlpatterns = [
    path('api/', include('api.urls')),
    path('intra/', include('intra_client.urls')),
    path('admins/', include('admins.urls')),
    path('admin/', admin.site.urls),
]
