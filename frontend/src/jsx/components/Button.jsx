export function Button({children, onClick, type, disabled}){
    return (
        <button type={type} 
                disabled={disabled} 
                onClick={onClick}
        >
            {children}
        </button>
    );
}