import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, Subject, timer } from 'rxjs';
import { retryWhen, delayWhen, tap } from 'rxjs/operators';

export enum WebSocketState {
  CONNECTING,
  CONNECTED,
  DISCONNECTED,
}

@Injectable({
  providedIn: 'root',
})
export class SocketService {
  private socket: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000; // 3秒
  private heartbeatInterval = 30000; // 30秒
  private heartbeatTimer: any;

  // 消息主题
  private readonly messageSubject = new Subject<any>();
  // 连接状态主题
  private readonly connectionStateSubject = new BehaviorSubject<WebSocketState>(
    WebSocketState.DISCONNECTED
  );

  // 公开的观察对象
  public readonly messages$ = this.messageSubject.asObservable();
  public readonly connectionState$ = this.connectionStateSubject.asObservable();

  constructor() {}

  // 连接到WebSocket服务器
  public connect(url: string = `ws://${window.location.host}/ws`): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      console.log('WebSocket已经连接');
      return;
    }

    this.connectionStateSubject.next(WebSocketState.CONNECTING);
    console.log('正在连接到WebSocket服务器...');

    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      console.log('WebSocket连接成功');
      this.connectionStateSubject.next(WebSocketState.CONNECTED);
      this.reconnectAttempts = 0;
      this.startHeartbeat();
    };

    this.socket.onmessage = (event) => {
      let data;
      try {
        data = JSON.parse(event.data);
      } catch (e) {
        data = event.data;
      }
      this.messageSubject.next(data);
    };

    this.socket.onclose = () => {
      console.log('WebSocket连接关闭');
      this.connectionStateSubject.next(WebSocketState.DISCONNECTED);
      this.stopHeartbeat();
      this.attemptReconnect(url);
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket错误:', error);
    };
  }

  // 发送消息
  public send(data: any): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      console.error('WebSocket未连接');
      return;
    }

    try {
      const message = typeof data === 'string' ? data : JSON.stringify(data);
      this.socket.send(message);
    } catch (error) {
      console.error('发送消息失败:', error);
    }
  }

  // 关闭连接
  public disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.stopHeartbeat();
    this.connectionStateSubject.next(WebSocketState.DISCONNECTED);
  }

  // 重连机制
  private attemptReconnect(url: string): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('达到最大重连次数');
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`
    );

    setTimeout(() => {
      this.connect(url);
    }, this.reconnectInterval);
  }

  // 启动心跳
  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      this.send({ type: 'ping' });
    }, this.heartbeatInterval);
  }

  // 停止心跳
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  // 获取当前连接状态
  public getConnectionState(): WebSocketState {
    return this.connectionStateSubject.value;
  }

  // 检查是否已连接
  public isConnected(): boolean {
    return this.connectionStateSubject.value === WebSocketState.CONNECTED;
  }
}
