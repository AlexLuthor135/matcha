import "./NavigationBar.css"
export default function NavigationBar(){
    return (
        <ul className="navigation-bar">
            <li><a className="active" href="/datingslider">Home</a></li>
            <li><a href="/userprofile">My Profile</a></li>
            <li><a href="#about">About</a></li>
        </ul>
    );
}