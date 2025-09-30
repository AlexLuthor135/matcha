import React from "react";

export function getCSRFToken() {
    const name = 'csrftoken';
    const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return match ? decodeURIComponent(match[2]) : null;
}
  