import type { UserSettings } from './usersettings';

export interface SignUpForm {
    email: string;
    password: string;
    wantReceiveUpdate: boolean;
    lang: string;
}

export interface LoginForm {
    email: string;
    password: string;
}

export interface User {
    id: string;
    email: string;
    CreatedAt: Date;
    LastSeen: Date;
    settings: UserSettings;
}

export default User;
