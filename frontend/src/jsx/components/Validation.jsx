export function passwordValidation(password, confirmPassword) {
    if(password !== confirmPassword){
        alert("NOT SAME PASSWORD!")
        return true;
    }
    if(password.length < 8){
        alert('PASSWORD MUST CONTAIN AT LEAST 8 CHARACTERS');
        return true;
    }
    if(!/[A-Z]/.test(password)){
        alert('NO UPPERCASE');
        return true;
    }
    if(!/[a-z]/.test(password)){
        alert('NO LOWERCASE');
        return true;
    }
    if(!/[0-9]/.test(password)){
        alert('NO NUMBERS');
        return true;
    }
    if(!/[!@#$%^&*]/.test(password)){
        alert('NO SPECIAL CHARACTER');
        return true;
    }
    if(/\s/.test(password)){
        alert('CANNOT CONTAINS SPACE');
        return true;
    }
    return false;
}
