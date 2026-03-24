export function InputBlock({type, value, onChange,placeholder,name,className}){
    return(
        
            <li>
                <input
                    type={type} 
                    value={value}
                    onChange={onChange}
                    placeholder={placeholder}
                    name={name}
                    className={className}
                />
            </li>
        
    );
}