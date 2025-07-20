import React from 'react';
import LoginButton from './LoginButton';
import SplashText from './SplashText';
import PatchNotes from './PatchNotes';

export default function LoginPage() {
  return (
    <div className="login-page">
      <SplashText />
      <LoginButton />
      <PatchNotes />
    </div>
  );
}

// function isAccessAllowed() {
//   const now = new Date();
//   const day = now.getDay();
//   const hour = now.getHours();

//   const isWeekday = day >= 1 && day <= 5;
//   const isWithinHours = hour >= 9 && hour < 17;

//   return isWeekday && isWithinHours;
// }

// export default function LoginPage() {
//   const accessAllowed = isAccessAllowed();

//   return (
//     <div className="login-page">
//       <SplashText />
//       {!accessAllowed && (
//         <p style={{ color: 'red', marginBottom: '1rem' }}>
//           Due to beta-testing, access is only allowed from 09:00 to 17:00 on weekdays.
//         </p>
//       )}
//       <LoginButton disabled={!accessAllowed} />
//     </div>
//   );
// }
