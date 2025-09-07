import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Bot, MessageCircle, Send, Sparkles, TrendingUp } from 'lucide-react';
import { stockService } from '../services/stockService';

interface SummaryResponse {
  summary: string;
  generated_at: string;
  tokens_used: number;
}

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

export const AIAssistant = () => {
  const [summary, setSummary] = useState<SummaryResponse | null>(null);
  const [loadingSummary, setLoadingSummary] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [sendingMessage, setSendingMessage] = useState(false);
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    loadSummary();
  }, []);

  const loadSummary = async () => {
    setLoadingSummary(true);
    try {
      const response = await fetch('http://localhost:8081/api/stocks/summary');
      if (response.ok) {
        const data = await response.json();
        setSummary(data);
      }
    } catch (error) {
      console.error('Failed to load AI summary:', error);
    } finally {
      setLoadingSummary(false);
    }
  };

  const sendMessage = async () => {
    if (!currentMessage.trim() || sendingMessage) return;

    const userMessage: ChatMessage = {
      role: 'user',
      content: currentMessage,
      timestamp: new Date()
    };

    setChatMessages(prev => [...prev, userMessage]);
    setCurrentMessage('');
    setSendingMessage(true);

    try {
      // Call your backend chat endpoint (you'll need to create this)
      const response = await fetch('http://localhost:8081/api/stocks/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: currentMessage })
      });

      if (response.ok) {
        const data = await response.json();
        const assistantMessage: ChatMessage = {
          role: 'assistant',
          content: data.response,
          timestamp: new Date()
        };
        setChatMessages(prev => [...prev, assistantMessage]);
      }
    } catch (error) {
      console.error('Failed to send chat message:', error);
      const errorMessage: ChatMessage = {
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date()
      };
      setChatMessages(prev => [...prev, errorMessage]);
    } finally {
      setSendingMessage(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* AI Market Summary */}
      <Card className="glass-card border border-border/50 animate-fade-in">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/20 animate-glow">
              <TrendingUp className="h-5 w-5 text-primary" />
            </div>
            AI Market Summary
          </CardTitle>
          <CardDescription>
            AI-powered analysis of current market trends and recommendations
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loadingSummary ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Bot className="h-4 w-4 animate-spin" />
              Generating AI analysis...
            </div>
          ) : summary ? (
            <div className="space-y-4">
              <p className="text-sm leading-relaxed">{summary.summary}</p>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Generated: {new Date(summary.generated_at).toLocaleString()}</span>
                <span>Tokens used: {summary.tokens_used}</span>
              </div>
              <Button 
                onClick={loadSummary} 
                variant="outline" 
                size="sm"
                disabled={loadingSummary}
              >
                <Sparkles className="h-4 w-4 mr-2" />
                Refresh Analysis
              </Button>
            </div>
          ) : (
            <p className="text-muted-foreground">Failed to load AI summary</p>
          )}
        </CardContent>
      </Card>

      {/* AI Chat Assistant */}
      <Card className="glass-card border border-border/50 animate-fade-in">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-success/20">
              <Bot className="h-5 w-5 text-success" />
            </div>
            Need Help?
          </CardTitle>
          <CardDescription>
            Ask our AI agent for personalized stock analysis and investment insights
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button className="w-full">
                <MessageCircle className="h-4 w-4 mr-2" />
                Chat with AI Assistant
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[80vh]">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <Bot className="h-5 w-5 text-primary" />
                  AI Stock Assistant
                </DialogTitle>
                <DialogDescription>
                  Ask questions about stocks, market trends, or get investment advice
                </DialogDescription>
              </DialogHeader>
              
              <div className="flex flex-col h-96">
                {/* Chat Messages */}
                <div className="flex-1 overflow-y-auto space-y-4 p-4 border rounded-lg bg-muted/20">
                  {chatMessages.length === 0 ? (
                    <div className="text-center text-muted-foreground">
                      <Bot className="h-8 w-8 mx-auto mb-2 opacity-50" />
                      <p>Hi! I'm your AI stock assistant. Ask me anything about the market!</p>
                    </div>
                  ) : (
                    chatMessages.map((message, index) => {
                      const messageAlignment = message.role === 'user' ? 'justify-end' : 'justify-start';
                      const messageStyle = message.role === 'user' 
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-background border';
                      
                      return (
                        <div
                          key={index}
                          className={`flex ${messageAlignment}`}
                        >
                          <div className={`max-w-[80%] p-3 rounded-lg ${messageStyle}`}>
                            <p className="text-sm">{message.content}</p>
                            <p className="text-xs opacity-70 mt-1">
                              {message.timestamp.toLocaleTimeString()}
                            </p>
                          </div>
                        </div>
                      );
                    })
                  )}
                  {sendingMessage && (
                    <div className="flex justify-start">
                      <div className="bg-background border p-3 rounded-lg">
                        <Bot className="h-4 w-4 animate-spin" />
                      </div>
                    </div>
                  )}
                </div>

                {/* Chat Input */}
                <div className="flex gap-2 mt-4">
                  <Input
                    value={currentMessage}
                    onChange={(e) => setCurrentMessage(e.target.value)}
                    placeholder="Ask about stocks, trends, or get investment advice..."
                    onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                    disabled={sendingMessage}
                  />
                  <Button 
                    onClick={sendMessage} 
                    disabled={!currentMessage.trim() || sendingMessage}
                  >
                    <Send className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </div>
  );
};