import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import axiosInstance from "../AxiosInstance";
import { useAuth } from "../AuthProvider";
import { useWebSocket } from "../WebSocketProvider";
import "./NotificationDropdown.css";

function normalizeNotification(notification) {
  if (!notification || typeof notification !== "object") {
    return null;
  }

  return {
    id: notification.id,
    title: String(notification.title ?? "Notification"),
    message: String(notification.message ?? ""),
    data: notification.data ?? null,
    readAt: notification.read_at ?? null,
    createdAt: notification.created_at ?? new Date().toISOString(),
  };
}

function formatNotificationTime(value) {
  if (!value) {
    return "";
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export default function NotificationDropdown() {
  const { isLoggedIn } = useAuth();
  const { subscribe, status } = useWebSocket();
  const [isOpen, setIsOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [error, setError] = useState("");
  const dropdownRef = useRef(null);

  const loadNotifications = useCallback(async () => {
    if (!isLoggedIn) {
      setNotifications([]);
      return;
    }

    try {
      const response = await axiosInstance.get("/backend/api/accounts/notifications");
      const items = Array.isArray(response.data?.notifications)
        ? response.data.notifications.map(normalizeNotification).filter(Boolean)
        : [];
      setNotifications(items);
      setError("");
    } catch (requestError) {
      console.log("NOTIFICATIONS LOAD ERROR", requestError);
      setError("Could not load notifications");
    }
  }, [isLoggedIn]);

  useEffect(() => {
    loadNotifications();
  }, [loadNotifications]);

  useEffect(() => {
    if (!isLoggedIn) {
      return undefined;
    }

    const intervalID = window.setInterval(loadNotifications, 10000);
    return () => window.clearInterval(intervalID);
  }, [isLoggedIn, loadNotifications]);

  useEffect(() => {
    if (isLoggedIn && status === "connected") {
      loadNotifications();
    }
  }, [isLoggedIn, loadNotifications, status]);

  useEffect(() => {
    if (!isLoggedIn) {
      return undefined;
    }

    return subscribe("notification", (payload) => {
      const notification = normalizeNotification(payload.notification);
      if (!notification) {
        return;
      }

      setNotifications((currentNotifications) => {
        if (currentNotifications.some((item) => item.id === notification.id)) {
          return currentNotifications;
        }
        return [notification, ...currentNotifications];
      });
    });
  }, [isLoggedIn, subscribe]);

  useEffect(() => {
    const handlePointerDown = (event) => {
      if (!dropdownRef.current?.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("pointerdown", handlePointerDown);
    return () => {
      document.removeEventListener("pointerdown", handlePointerDown);
    };
  }, []);

  const unreadCount = useMemo(
    () => notifications.filter((notification) => !notification.readAt).length,
    [notifications]
  );

  const markNotificationRead = async (notificationID) => {
    setNotifications((currentNotifications) => currentNotifications.map((notification) => (
      notification.id === notificationID
        ? { ...notification, readAt: notification.readAt || new Date().toISOString() }
        : notification
    )));

    try {
      await axiosInstance.patch(`/backend/api/accounts/notifications/${notificationID}/read`);
    } catch (requestError) {
      console.log("NOTIFICATION READ ERROR", requestError);
    }
  };

  if (!isLoggedIn) {
    return null;
  }

  return (
    <div className="notification-dropdown" ref={dropdownRef}>
      <button
        className="notification-button"
        type="button"
        aria-expanded={isOpen}
        aria-haspopup="true"
        onClick={() => setIsOpen((currentIsOpen) => !currentIsOpen)}
      >
        Notifications
        {unreadCount > 0 ? <span className="notification-badge">{unreadCount}</span> : null}
      </button>

      {isOpen ? (
        <div className="notification-menu" role="menu">
          <div className="notification-menu-header">
            <h3>Notifications</h3>
          </div>

          {error ? <p className="notification-status">{error}</p> : null}

          {!error && notifications.length === 0 ? (
            <p className="notification-status">No notifications yet</p>
          ) : null}

          <div className="notification-list">
            {notifications.map((notification) => (
              <button
                key={notification.id}
                className={`notification-item${notification.readAt ? "" : " is-unread"}`}
                type="button"
                role="menuitem"
                onClick={() => markNotificationRead(notification.id)}
              >
                <span className="notification-title">{notification.title}</span>
                <span className="notification-message">{notification.message}</span>
                <time className="notification-time">
                  {formatNotificationTime(notification.createdAt)}
                </time>
              </button>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}
