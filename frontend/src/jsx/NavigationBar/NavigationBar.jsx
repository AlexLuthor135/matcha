import "./NavigationBar.css"
import AgeFilterButton from "./AgeFilterButton";
import NotificationDropdown from "../components/NotificationDropdown";

export default function NavigationBar(){
    return (
        <ul className="navigation-bar">
            <li><a className="active" href="/datingslider">Home</a></li>
            <li><a href="/chat">Chat</a></li>
            <li><a href="/userprofile">My Profile</a></li>
            <li><a href="#about">About</a></li>
            <li className="navigation-filter"><AgeFilterButton /></li>
            <li className="navigation-notifications"><NotificationDropdown /></li>
        </ul>
    );
}
