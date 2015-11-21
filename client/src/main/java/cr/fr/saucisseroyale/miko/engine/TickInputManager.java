package cr.fr.saucisseroyale.miko.engine;

import cr.fr.saucisseroyale.miko.util.Triplet;

import java.awt.Point;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * Un gestionnaire des inputs à chaque tick de jeu.
 *
 */
class TickInputManager {

  private Map<Long, TickInput> map;
  private long firstTick = Long.MAX_VALUE;

  public void addInput(long tick, List<Triplet<Boolean, Integer, Point>> eventList) {
    TickInput previousInput = getInput(tick - 1);
    TickInput newInput = new TickInput(previousInput, eventList);
    map.put(tick, newInput);
  }

  public TickInput getInput(long tick) {
    if (tick < firstTick) {
      return null;
    }
    TickInput tickInput = map.get(tick);
    if (tickInput == null) {
      TickInput previousInput = getInput(tick - 1);
      TickInput newInput = TickInput.getNextFrom(previousInput);
      map.put(tick, newInput);
      return newInput;
    } else {
      return tickInput;
    }
  }

  public void disposeUntilTick(long tick) {
    Set<Map.Entry<Long, TickInput>> entrySet = map.entrySet();
    Iterator<Map.Entry<Long, TickInput>> iterator = entrySet.iterator();
    while (iterator.hasNext()) {
      Map.Entry<Long, TickInput> entry = iterator.next();
      if (entry.getKey() > tick) {
        firstTick = entry.getKey();
        break;
      }
      iterator.remove();
    }
  }

}
