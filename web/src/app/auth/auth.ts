import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private apiUrl = '/api/auth';

  constructor(private http: HttpClient) { }
  login(credentials: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/login`, credentials).pipe(
      tap((response: any) => {
        if (response && response.access_token) {
          localStorage.setItem('auth_token', response.access_token);
        }
      })
    );
  }
  register(userData: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/register`, userData);
  }

  logout(): void {
    localStorage.removeItem('auth_token');
  }

  public get isLoggedIn(): boolean {
    return localStorage.getItem('auth_token') !== null;
  }

  getProfile(): Observable<any> {
    return this.http.get(`${this.apiUrl}/me`); 
  }
}