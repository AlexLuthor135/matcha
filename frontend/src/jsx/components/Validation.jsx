export function passwordValidation(password, confirmPassword) {
    if(password !== confirmPassword){
        alert("NOT SAME PASSWORD!")
        return true;
    }
    if(!/[A-Z]/.test(password)){
        alert('NO UPPERCASE');
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