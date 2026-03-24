import { Button } from "../components/Button";
import axiosInstance from "../AxiosInstance";

export default function UserProfile(){
    const handleLogout= async (e) => {
        e.preventDefault();
        try{
            const response = await axiosInstance.post('/backend/api/accounts/logout/', {
            })
            console.log("LOGOUT COMLETED", response.status)
        }
        catch(error){
            console.log("LOGOUT ERROR", error);
        }
        finally{

        }
    }
    return(
        <div>
            <h2>SUCCESS</h2>
            <Button onClick={handleLogout}>LOGOUT</Button>
        </div>
    );
}