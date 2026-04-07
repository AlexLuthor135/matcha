export function Button({children, onClick, type = "button", disabled, iconSrc, iconAlt, ariaLabel, className}){
    const buttonClassName = [className, iconSrc ? "icon-button" : null].filter(Boolean).join(" ");

    return (
        <button type={type} 
                disabled={disabled} 
                onClick={onClick}
                aria-label={ariaLabel}
                className={buttonClassName}
        >
            {iconSrc ? <img src={iconSrc} alt={iconAlt ?? ""} /> : children}
        </button>
    );
}