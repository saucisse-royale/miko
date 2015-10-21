package cr.fr.saucisseroyale.miko.network;

import java.io.IOException;
import java.io.InputStream;
import java.net.Socket;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.cert.CertificateException;
import java.util.Queue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.LinkedBlockingQueue;

import javax.net.ssl.SSLContext;
import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.TrustManagerFactory;

/**
 * Client de connexion à un serveur Miko fonctionnant sur la couche des messages.
 *
 * @see FutureInputMessage
 * @see FutureOutputMessage
 *
 */
public class NetworkClient {

  private Socket socket;
  private ReceiverThread receiverThread;
  private SenderThread senderThread;
  private Queue<FutureInputMessage> inputMessages = new ConcurrentLinkedQueue<>();
  private BlockingQueue<FutureOutputMessage> outputMessages = new LinkedBlockingQueue<>();

  /**
   * Se connecte au serveur spécifié et démarre les threads d'envoi et de réception des messages.
   *
   * @param address L'adresse du serveur auquel se connecter (IP ou nom d'hôte).
   * @param port Le port auquel se connecter.
   * @throws IOException S'il y a des erreurs quelconques d'IO lors de la connexion.
   */
  public void connect(String address, int port) throws IOException {
    // TODO accepter seulement un certain certificat ?
    SSLSocketFactory socketFactory;
    try {
      socketFactory = createSocketFactory();
    } catch (Exception e) {
      // TODO log cette erreur
      throw new RuntimeException(e);
    }
    socket = socketFactory.createSocket(address, port);
    socket.setTcpNoDelay(true);
    socket.setTrafficClass(0x10); // LOWDELAY
    receiverThread = new ReceiverThread(socket.getInputStream(), inputMessages);
    senderThread = new SenderThread(socket.getOutputStream(), outputMessages);
  }


  /**
   * Termine définitivement la connection et l'envoi et la réception des messages.
   *
   * @throws IOException S'il y a des erreurs quelconques d'IO lors de l'extinction.
   */
  public void exit() throws IOException {
    receiverThread.interrupt();
    senderThread.interrupt();
    socket.close();
  }

  /**
   * Renvoie un message de la liste de réception des messages.
   *
   * @return Le message reçu le plus ancien, ou <code>null</code> si aucun message n'est en attente.
   *
   */
  public FutureInputMessage getMessage() {
    return inputMessages.poll();
  }

  /**
   * Ajoute un message à la liste d'envoi des messages.
   *
   * @param fom Le message à envoyer.
   */
  public void putMessage(FutureOutputMessage fom) {
    outputMessages.add(fom);
  }

  private static SSLSocketFactory createSocketFactory() throws KeyStoreException, IOException,
      NoSuchAlgorithmException, CertificateException, KeyManagementException {
    KeyStore keyStore = KeyStore.getInstance(KeyStore.getDefaultType());
    try (InputStream keyStoreStream = NetworkClient.class.getResourceAsStream("/keystore")) {
      keyStore.load(keyStoreStream, null);
    }
    TrustManagerFactory tmf =
        TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
    tmf.init(keyStore);
    SSLContext context = SSLContext.getInstance("TLS");
    context.init(null, tmf.getTrustManagers(), null);
    return context.getSocketFactory();
  }

}
