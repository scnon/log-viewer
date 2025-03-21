import { Component, OnDestroy, OnInit } from '@angular/core';
import { SocketService, WebSocketState } from './socket.service';
import { LogViewerComponent } from './log-viewer/log-viewer.component';
import { Subscription } from 'rxjs';
import { LineChange, Log } from '../types';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzButtonModule } from 'ng-zorro-antd/button';

export interface Message {
  type: string;
  data: any;
}

@Component({
  selector: 'app-root',
  imports: [
    LogViewerComponent,
    NzTableModule,
    NzInputModule,
    NzInputModule,
    NzButtonModule,
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnInit, OnDestroy {
  title = 'ui';
  logs: Log[] = [];
  lineChanges: LineChange[] = [];
  private subscriptions: Subscription[] = [];
  connectionStatus = '未连接';
  constructor(private socketService: SocketService) {}

  ngOnInit() {
    // 订阅连接状态
    this.subscriptions.push(
      this.socketService.connectionState$.subscribe((state) => {
        switch (state) {
          case WebSocketState.CONNECTED:
            this.connectionStatus = '已连接';
            this.socketService.send({
              type: 'get_info',
            });
            break;
          case WebSocketState.CONNECTING:
            this.connectionStatus = '连接中...';
            break;
          case WebSocketState.DISCONNECTED:
            this.connectionStatus = '未连接';
            break;
        }
      })
    );

    // 订阅消息
    this.subscriptions.push(
      this.socketService.messages$.subscribe((message) => {
        switch (message.type) {
          case 'pong':
            break;
          case 'info':
            this.init(message.data);
            break;
          case 'file_content':
            this.initLog(message.data);
            break;
          case 'log':
            this.logs.push(message.data);
            for (const log of message.data.line_changes) {
              this.lineChanges.push(log);
            }
            break;
          default:
            console.log('未知消息:', message);
            break;
        }
      })
    );

    // 连接到WebSocket服务器
    this.socketService.connect('ws://localhost:8081/ws');
  }

  ngOnDestroy() {
    // 清理订阅
    this.subscriptions.forEach((sub) => sub.unsubscribe());
    // 断开连接
    this.socketService.disconnect();
  }

  init(data: any) {
    this.socketService.send({
      type: 'get_file_content',
      data: data.path,
    });
  }

  initLog(data: string) {
    const lines = data.split('\n');
    this.logs = [
      {
        path: 'test.log',
        modified: false,
        content: '',
        prev_content: '',
        line_changes: lines.map((line) => {
          return {
            type: 'log',
            old_line: 0,
            new_line: 0,
            old_text: '',
            new_text: line,
          };
        }),
      },
    ];
    this.lineChanges = this.logs[0].line_changes;
    console.log(this.logs);
  }
}
